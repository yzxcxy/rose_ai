package svc

import (
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"rose/internal/agent"
	"rose/internal/config"
	cryptUtil "rose/pkg/crypt"
)

type ServiceContext struct {
	Config config.Config
	Mysql  sqlx.SqlConn
	Redis  *redis.Client
	Cipher cryptUtil.Cipher
	Agent  *agent.Agent
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysql := sqlx.NewMysql(c.MySQL.DSN)

	rd := redis.NewClient(&redis.Options{
		Addr: c.Redis.RedisHost,
	})

	a, err := agent.NewAgent(&c)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		Mysql:  mysql,
		Redis:  rd,
		Cipher: cryptUtil.NewBcryptCipher(10),
		Agent:  a,
	}
}
