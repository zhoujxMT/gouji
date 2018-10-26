/*
比赛redis初始化
*/

package engine

import (
	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"
	"kelei.com/utils/lua"
)

const (
	expireSecond_Short     = 10
	expireSecond           = 10
	expireSecond_Long      = 60
	expireSecond_Long_Long = 60 * 60
)

var (
	luaScript string
)

func redis_Init() {
	//加载脚本
	content, err := lua.LoadLuaFile("./engine/lua/lua.lua")
	logger.CheckFatal(err)
	gameRds := getGameRds()
	defer gameRds.Close()
	luaScript, err = redis.String(gameRds.Do("script", "load", content))
	logger.CheckFatal(err)
}
