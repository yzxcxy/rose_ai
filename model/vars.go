package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

var ErrNotFound = sqlx.ErrNotFound

type ListToDos struct {
	// 用户 id 是必要的
	UserId   int64
	Page     int64 // 页码
	PageSize int64
	Status   []string
	SortBy   []string // 排序字段
	Priority []string

	SortOrder  []string // 排序方式，asc 或 desc
	MinDueDate time.Time
	MaxDueDate time.Time
	Search     string // 搜索关键字
	StartDate  time.Time
	EndDate    time.Time // 任务的创建时间
}
