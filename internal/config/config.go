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

	DeepSeek struct {
		BaseURL   string // DeepSeek API 基础 URL
		Token     string // DeepSeek API 访问令牌
		Model     string // DeepSeek 模型名称
		MaxTokens int    // DeepSeek 最大令牌数
	}

	Doubao struct {
		APIKey string // Doubao API Key
		Model  string // Doubao 模型名称
		AK     string
		SK     string
	}

	VikingDB struct {
		Host              string // VikingDB 主机地址
		Region            string // VikingDB 区域
		AK                string // VikingDB 访问密钥 AK
		SK                string // VikingDB 访问密钥 SK
		Scheme            string // VikingDB 连接协议 (http 或 https)
		Collection        string // VikingDB 集合名称
		UseBuiltin        bool   // 是否使用 VikingDB 内置向量化方法
		ConnectionTimeout int64  // 连接超时时间，单位为秒
		Chunk             int    // 一次性插入的大小
	}

	Milvus struct {
		Host   string // Milvus 主机地址
		DBName string // Milvus 数据库名称
	}
}
