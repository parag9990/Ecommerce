package repository

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisClientConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type RedisClient struct {
	addr         string
	password     string
	db           int
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewRedisClient(cfg RedisClientConfig) (*RedisClient, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, errors.New("redis address is required")
	}
	if cfg.DB < 0 {
		return nil, errors.New("redis db cannot be negative")
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 2 * time.Second
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 2 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 2 * time.Second
	}

	return &RedisClient{
		addr:         strings.TrimSpace(cfg.Addr),
		password:     cfg.Password,
		db:           cfg.DB,
		dialTimeout:  cfg.DialTimeout,
		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,
	}, nil
}

func (c *RedisClient) Ping(ctx context.Context) error {
	value, err := c.do(ctx, "PING")
	if err != nil {
		return err
	}
	if value.stringValue() != "PONG" {
		return errors.New("unexpected redis ping response")
	}
	return nil
}

func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	value, err := c.do(ctx, "EXISTS", key)
	if err != nil {
		return false, err
	}
	count, err := value.intValue()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, bool, error) {
	if strings.TrimSpace(key) == "" {
		return "", false, errors.New("redis key is required")
	}
	value, err := c.do(ctx, "GET", key)
	if err != nil {
		return "", false, err
	}
	if value.isNil {
		return "", false, nil
	}
	return value.stringValue(), true, nil
}

func (c *RedisClient) SetEX(ctx context.Context, key string, value string, ttl time.Duration) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("redis key is required")
	}
	if ttl <= 0 {
		return errors.New("redis set ttl must be greater than zero")
	}
	resp, err := c.do(ctx, "SET", key, value, "EX", strconv.Itoa(secondsCeil(ttl)))
	if err != nil {
		return err
	}
	if resp.stringValue() != "OK" {
		return errors.New("unexpected redis set response")
	}
	return nil
}

func (c *RedisClient) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, errors.New("redis set ttl must be greater than zero")
	}
	resp, err := c.do(ctx, "SET", key, value, "NX", "EX", strconv.Itoa(secondsCeil(ttl)))
	if err != nil {
		return false, err
	}
	if resp.isNil {
		return false, nil
	}
	return resp.stringValue() == "OK", nil
}

func (c *RedisClient) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("redis key is required")
	}
	resp, err := c.do(ctx, "DEL", key)
	if err != nil {
		return err
	}
	_, err = resp.intValue()
	return err
}

func (c *RedisClient) do(ctx context.Context, args ...string) (redisValue, error) {
	if c == nil {
		return redisValue{}, errors.New("redis client is not initialized")
	}
	if len(args) == 0 {
		return redisValue{}, errors.New("redis command is required")
	}

	dialer := net.Dialer{Timeout: c.dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return redisValue{}, fmt.Errorf("dial redis: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	if c.password != "" {
		if _, err := c.execute(ctx, conn, reader, "AUTH", c.password); err != nil {
			return redisValue{}, fmt.Errorf("redis auth: %w", err)
		}
	}
	if c.db > 0 {
		if _, err := c.execute(ctx, conn, reader, "SELECT", strconv.Itoa(c.db)); err != nil {
			return redisValue{}, fmt.Errorf("redis select db: %w", err)
		}
	}
	return c.execute(ctx, conn, reader, args...)
}

func (c *RedisClient) execute(ctx context.Context, conn net.Conn, reader *bufio.Reader, args ...string) (redisValue, error) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	if _, err := conn.Write(encodeRedisCommand(args...)); err != nil {
		return redisValue{}, fmt.Errorf("write redis command: %w", err)
	}

	if _, ok := ctx.Deadline(); !ok {
		_ = conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
	value, err := readRedisValue(reader)
	if err != nil {
		return redisValue{}, err
	}
	return value, nil
}

func encodeRedisCommand(args ...string) []byte {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(args)))
	builder.WriteString("\r\n")
	for _, arg := range args {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(arg)))
		builder.WriteString("\r\n")
		builder.WriteString(arg)
		builder.WriteString("\r\n")
	}
	return []byte(builder.String())
}

type redisValue struct {
	kind  byte
	value string
	i64   int64
	isNil bool
}

func (v redisValue) stringValue() string {
	if v.kind == ':' {
		return strconv.FormatInt(v.i64, 10)
	}
	return v.value
}

func (v redisValue) intValue() (int64, error) {
	if v.kind == ':' {
		return v.i64, nil
	}
	if v.value == "" {
		return 0, errors.New("redis value is not an integer")
	}
	return strconv.ParseInt(v.value, 10, 64)
}

func readRedisValue(reader *bufio.Reader) (redisValue, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return redisValue{}, fmt.Errorf("read redis response: %w", err)
	}

	switch prefix {
	case '+':
		line, err := readRedisLine(reader)
		return redisValue{kind: prefix, value: line}, err
	case '-':
		line, err := readRedisLine(reader)
		if err != nil {
			return redisValue{}, err
		}
		return redisValue{}, errors.New("redis error: " + line)
	case ':':
		line, err := readRedisLine(reader)
		if err != nil {
			return redisValue{}, err
		}
		value, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return redisValue{}, fmt.Errorf("parse redis integer: %w", err)
		}
		return redisValue{kind: prefix, i64: value}, nil
	case '$':
		line, err := readRedisLine(reader)
		if err != nil {
			return redisValue{}, err
		}
		length, err := strconv.Atoi(line)
		if err != nil {
			return redisValue{}, fmt.Errorf("parse redis bulk length: %w", err)
		}
		if length == -1 {
			return redisValue{kind: prefix, isNil: true}, nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return redisValue{}, fmt.Errorf("read redis bulk value: %w", err)
		}
		return redisValue{kind: prefix, value: string(buf[:length])}, nil
	default:
		return redisValue{}, fmt.Errorf("unsupported redis response prefix %q", prefix)
	}
}

func readRedisLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read redis line: %w", err)
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func secondsCeil(d time.Duration) int {
	seconds := int(d / time.Second)
	if d%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		return 1
	}
	return seconds
}
