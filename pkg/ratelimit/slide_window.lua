-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber( ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口的起始时间
local min = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
-- 获取当前窗口内的请求数量
-- local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
local cnt = redis.call('ZCOUNT', key, min, '+inf')
-- 如果当前请求数大于阈值执行限流，否则请求存进redis
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 把 score 和 member 都设置成 now
    redis.call('ZADD', key, now, now)
    -- 设置过期时间
    redis.call('PEXPIRE', key, window)
    return "false"
end