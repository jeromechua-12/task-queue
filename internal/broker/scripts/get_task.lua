-- Script that dequeues a task id from Redis Sorted Set
-- Use the task id to get the Task object from Redis Hash 
-- Returns Task object. Returns nil if no task in Redis Sorted Set
--
-- Takes in 2 keys and 0 args.
-- KEYS[1]: Redis Sorted Set storing task id ready to be dequeued
-- KEYS[2]: Redis Hash that maps task id to Task in serialised JSON

if #KEYS~= 2 then
    local key_err = "KEY ERROR: invalid number of keys. Expected 2; got " .. #KEYS
    error(key_err)
end

if #ARGV > 0 then
    local args_err = "ARGS ERROR: invalid number of keys. Expected 0; got " .. #ARGV
    error(args_err)
end

local sortedSet = KEYS[1]
local hash = KEYS[2]

-- pop task id from Sorted Set
local task_ids = redis.pcall("ZPOPMIN", sortedSet, 1)
if task_ids["err"] ~= nil then
    local pop_err = "ZPOPMIN ERROR: " .. task_ids["err"]
    error(pop_err)
end

-- check if there are any popped elements
if #task_ids == 0 then
    return nil
end

-- get task details from Hash
local task_id = task_ids[1]

local task = redis.pcall("HGET", hash, task_id)
if task["err"] ~= nil then
    local get_err = "HGET ERROR: " .. task["err"]
    error(get_err)
end

return task
