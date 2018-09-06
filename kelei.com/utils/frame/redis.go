package frame

import (
	redis_ "github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"
	"kelei.com/utils/redis"
)

var (
	rdses = make(map[string]*redis.RDS)
)

type Redis struct {
	RedisDSNs []*redis.RedisDSN
}

func loadRedis() {
	if args.Redis == nil {
		return
	}
	redisDSNs := args.Redis.RedisDSNs
	for _, redisDSN := range redisDSNs {
		logger.Infof("[连接Redis %s]", redisDSN.Addr)
		rdses[redisDSN.Name] = redis.NewPool(redisDSN)
		GetRedis().Do("flushall")
	}
}

//获得redis
func GetRedis(args ...string) redis_.Conn {
	var rds *redis.RDS
	if len(args) > 0 {
		rdsName := args[0]
		for _, rds_ := range rdses {
			if rds_.RedisDSN.Name == rdsName {
				rds = rds_
				break
			}
		}
	} else {
		for _, rds_ := range rdses {
			rds = rds_
			break
		}
	}
	return rds.RedisPool.Get()
}
