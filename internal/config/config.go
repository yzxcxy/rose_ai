package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Auth struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}

	MySQL struct {
		DSN string // 数据库连接字符串
	}

	Redis struct {
		RedisHost string // Redis服务器地址
	}

	SnowFlake struct {
		WorkerId  int64  // 雪花算法 Worker ID
		StartTime string // 雪花算法起始时间
	}
}
