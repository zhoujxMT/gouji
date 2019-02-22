package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"
)

type RDS struct {
	RedisDSN  *RedisDSN
	RedisPool *redis.Pool
}

type RedisDSN struct {
	Name     string
	Addr     string
	PassWord string
}

//初始化一个pool
func NewPool(redisDSN *RedisDSN) *RDS {
	addr := redisDSN.Addr
	password := redisDSN.PassWord
	redisPool := &redis.Pool{
		MaxIdle:     200,
		MaxActive:   1024,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				logger.CheckFatal(err, "redis.Dial")
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					logger.CheckFatal(err, "redis.Dial")
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return &RDS{redisDSN, redisPool}
}
