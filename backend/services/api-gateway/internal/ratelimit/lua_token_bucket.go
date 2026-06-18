package ratelimit

const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_per_ms = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local ttl_ms = tonumber(ARGV[5])

local bucket = redis.call("HMGET", key, "tokens", "updated_at")
local tokens = tonumber(bucket[1])
local updated_at = tonumber(bucket[2])

if tokens == nil then
  tokens = capacity
  updated_at = now_ms
end

local elapsed = math.max(0, now_ms - updated_at)
local refill = elapsed * refill_per_ms
tokens = math.min(capacity, tokens + refill)

local allowed = 0
local retry_after_ms = 0

if tokens >= cost then
  allowed = 1
  tokens = tokens - cost
else
  retry_after_ms = math.ceil((cost - tokens) / refill_per_ms)
end

local reset_after_ms = math.ceil((capacity - tokens) / refill_per_ms)
redis.call("HSET", key, "tokens", tokens, "updated_at", now_ms)
redis.call("PEXPIRE", key, ttl_ms)

return {allowed, math.floor(tokens), retry_after_ms, reset_after_ms}
`
