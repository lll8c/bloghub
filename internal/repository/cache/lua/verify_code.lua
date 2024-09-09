-- 验证码在redis上的key
local key = KEYS[1]

local cntKey = key..":cnt"
-- 用户输入的验证码
local expectedCode = ARGV[1]

local cnt = tonumber(redis.call("get", cntKey))
local code = redis.call("get", key)

if cnt == nil or cnt <= 0 then
--    验证次数耗尽了
    return -1
end

if code == expectedCode then
    -- 验证码用完不能再用了
    -- 剩余可用次数置为-1
    redis.call("set", cntKey, -1)
    return 0
else
    -- 可验证次数减1
    redis.call("decr", cntKey)
    -- 不相等，用户输错了
    return -2
end