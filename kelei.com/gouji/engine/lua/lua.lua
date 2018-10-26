
local expireSecond = 10

local split = function(input, delimiter,plain)
    input = tostring(input)
    delimiter = tostring(delimiter)
    if plain==nil then
        plain=true
    end
    if (delimiter=='') then return false end
    local pos,arr = 0, {}
    -- for each divider found
    for st,sp in function() return string.find(input, delimiter, pos, plain) end do
        table.insert(arr, string.sub(input, pos, st - 1))
        pos = sp + 1
    end
    table.insert(arr, string.sub(input, pos))
    return arr
end

--加载表balancedata
local balancedata = function(content)
	local rows = split(content,"|")
	local infos = ""
	local key = ""
	local roomType, balanceType, ingot, integral
	for i,row in ipairs(rows) do
		infos = split(row,",")
		roomType, balanceType, ingot, integral = infos[1], infos[2], infos[3], infos[4]
		key = "roomT:"..roomType..":blceT:"..balanceType
		redis.call("hmset", key, "ingot", ingot, "integral", integral)
		redis.call("expire", key, expireSecond)
	end
end

for i,funcName in ipairs(KEYS) do
	local content = ARGV[i]
	if funcName=="balancedata" then
		balancedata(content)
	end
end