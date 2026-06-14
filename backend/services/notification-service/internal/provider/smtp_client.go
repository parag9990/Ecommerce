package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

type SMTPClientConfig struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
	TLSMode  string
	Timeout  time.Duration
}

type SMTPClient struct {
	config  SMTPClientConfig
	address string
	from    string
}

func NewSMTPClient(config SMTPClientConfig) (*SMTPClient, error) {
	config.Host = strings.TrimSpace(config.Host)
	config.From = strings.TrimSpace(config.From)
	config.Username = strings.TrimSpace(config.Username)
	config.Password = strings.TrimSpace(config.Password)
	config.TLSMode = strings.ToLower(strings.TrimSpace(config.TLSMode))
	if config.Host == "" || config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("SMTP host and valid port are required")
	}
	from, err := mail.ParseAddress(config.From)
	if err != nil || from.Address != config.From {
		return nil, errors.New("SMTP from must be a valid email address")
	}
	switch config.TLSMode {
	case "tls", "starttls", "none":
	default:
		return nil, errors.New("SMTP TLS mode must be tls, starttls, or none")
	}
	if (config.Username == "") != (config.Password == "") {
		return nil, errors.New("SMTP username and password must be configured together")
	}
	if config.Username != "" && config.TLSMode == "none" {
		return nil, errors.New("SMTP authentication requires TLS")
	}
	if config.Timeout <= 0 {
		return nil, errors.New("SMTP timeout must be greater than zero")
	}
	return &SMTPClient{
		config:  config,
		address: net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
		from:    from.Address,
	}, nil
}

func (c *SMTPClient) Deliver(ctx context.Context, to, subject, textBody, htmlBody string) (string, error) {
	if c == nil {
		return "", ErrProviderUnavailable
	}
	if ctx == nil {
		return "", ErrProviderRejected
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	to = strings.TrimSpace(to)
	address, err := mail.ParseAddress(to)
	if err != nil || address.Address != to || strings.ContainsAny(subject, "\r\n") {
		return "", ErrProviderRejected
	}

	messageID, err := smtpMessageID()
	if err != nil {
		return "", ErrProviderUnavailable
	}
	body, err := smtpMessage(c.from, address.Address, subject, textBody, htmlBody, messageID, c.config.Host)
	if err != nil {
		return "", ErrProviderUnavailable
	}

	deliveryCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	if err := c.send(deliveryCtx, address.Address, body); err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", ErrProviderUnavailable
	}
	return messageID, nil
}

func (c *SMTPClient) send(ctx context.Context, to string, message []byte) error {
	connection, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	if deadline, ok := ctx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			return err
		}
	}
	stopCancellation := make(chan struct{})
	defer close(stopCancellation)
	go func() {
		select {
		case <-ctx.Done():
			_ = connection.Close()
		case <-stopCancellation:
		}
	}()

	client, err := smtp.NewClient(connection, c.config.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if c.config.TLSMode == "starttls" {
		if err := client.StartTLS(c.tlsConfig()); err != nil {
			return err
		}
	}
	if c.config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)); err != nil {
			return err
		}
	}
	if err := client.Mail(c.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	// DATA acknowledgement is the acceptance point; a failed QUIT must not
	// cause an already accepted OTP to be sent again by the caller.
	_ = client.Quit()
	return nil
}

func (c *SMTPClient) dial(ctx context.Context) (net.Conn, error) {
	if c.config.TLSMode == "tls" {
		dialer := tls.Dialer{Config: c.tlsConfig()}
		return dialer.DialContext(ctx, "tcp", c.address)
	}
	return (&net.Dialer{}).DialContext(ctx, "tcp", c.address)
}

func (c *SMTPClient) tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: c.config.Host,
	}
}

func smtpMessage(from, to, subject, textBody, htmlBody, id, host string) ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(&output, "From: %s\r\n", from)
	fmt.Fprintf(&output, "To: %s\r\n", to)
	fmt.Fprintf(&output, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	fmt.Fprintf(&output, "Message-ID: <%s@%s>\r\n", id, host)
	output.WriteString("MIME-Version: 1.0\r\n")
	if htmlBody == "" {
		output.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		output.WriteString(textBody)
		return output.Bytes(), nil
	}

	multi := multipart.NewWriter(&output)
	fmt.Fprintf(&output, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", multi.Boundary())
	textHeader := textproto.MIMEHeader{"Content-Type": []string{"text/plain; charset=UTF-8"}}
	text, err := multi.CreatePart(textHeader)
	if err != nil {
		return nil, err
	}
	if _, err := text.Write([]byte(textBody)); err != nil {
		return nil, err
	}
	htmlHeader := textproto.MIMEHeader{"Content-Type": []string{"text/html; charset=UTF-8"}}
	html, err := multi.CreatePart(htmlHeader)
	if err != nil {
		return nil, err
	}
	if _, err := html.Write([]byte(htmlBody)); err != nil {
		return nil, err
	}
	if err := multi.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func smtpMessageID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
