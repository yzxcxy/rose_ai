package svc

import (
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"rose/internal/config"
	cryptUtil "rose/pkg/crypt"
)

type ServiceContext struct {
	Config config.Config
	Mysql  sqlx.SqlConn
	Redis  *redis.Client
	Cipher cryptUtil.Cipher
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysql := sqlx.NewMysql(c.MySQL.DSN)

	rd := redis.NewClient(&redis.Options{
		Addr: c.Redis.RedisHost,
	})

	return &ServiceContext{
		Config: c,
		Mysql:  mysql,
		Redis:  rd,
		Cipher: cryptUtil.NewBcryptCipher(10),
	}
}
