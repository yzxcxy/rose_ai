package model

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"rose/internal/types"
	"strings"
)

var _ TodosModel = (*customTodosModel)(nil)

type (
	// TodosModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTodosModel.
	TodosModel interface {
		todosModel
		withSession(session sqlx.Session) TodosModel
		ListToDos(ctx context.Context, list ListToDos) ([]*Todos, error)
	}

	customTodosModel struct {
		*defaultTodosModel
	}
)

// NewTodosModel returns a model for the database table.
func NewTodosModel(conn sqlx.SqlConn) TodosModel {
	return &customTodosModel{
		defaultTodosModel: newTodosModel(conn),
	}
}

func (m *customTodosModel) withSession(session sqlx.Session) TodosModel {
	return NewTodosModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTodosModel) ListToDos(ctx context.Context, list ListToDos) ([]*Todos, error) {
	// 动态拼接sql
	if list.UserId <= 0 {
		return nil, types.GetError(types.ErrorUserNotFound)
	}

	// 基础查询
	query := sq.Select(
		"todo_id", "user_id", "name", "description",
		"status", "priority", "due_date", "created_at",
		"updated_at", "is_deleted",
	).From("todos").Where(sq.Eq{"user_id": list.UserId, "is_deleted": 0})

	if list.Status != nil {
		// 传入的是切片，会自动形成 IN 查询
		query = query.Where(sq.Eq{"status": list.Status})
	}

	// 关键字搜索
	if strings.TrimSpace(list.Search) != "" {
		like := "%" + list.Search + "%"
		query = query.Where(
			sq.Or{
				sq.Like{"name": like},
				sq.Like{"description": like},
			},
		)
	}

	// 优先级过滤
	if list.Priority != nil {
		// 传入的是切片，会自动形成 IN 查询
		query = query.Where(sq.Eq{"priority": list.Priority})
	}

	// 截止时间区间
	if !list.MinDueDate.IsZero() {
		query = query.Where(sq.GtOrEq{"due_date": list.MinDueDate})
	}
	if !list.MaxDueDate.IsZero() {
		query = query.Where(sq.LtOrEq{"due_date": list.MaxDueDate})
	}

	// 创建时间区间
	if !list.StartDate.IsZero() {
		query = query.Where(sq.GtOrEq{"created_at": list.StartDate})
	}
	if !list.EndDate.IsZero() {
		query = query.Where(sq.LtOrEq{"created_at": list.EndDate})
	}

	// 排序
	// 1. 统一决定排序方向
	sortOrder := "DESC"
	if list.SortOrder != "" {
		if strings.ToUpper(list.SortOrder) == "ASC" {
			sortOrder = "ASC"
		}
	}

	// 2. 允许的字段白名单
	allowed := map[string]struct{}{
		"id": {}, "todo_id": {}, "name": {}, "status": {}, "priority": {},
		"due_date": {}, "created_at": {}, "updated_at": {},
	}

	// 3. 遍历 SortBy 切片，过滤并拼接
	var orderBys []string
	for _, col := range list.SortBy {
		col = strings.ToLower(col)
		if _, ok := allowed[col]; ok {
			orderBys = append(orderBys, fmt.Sprintf("%s %s", col, sortOrder))
		}
	}

	// 4. 兜底
	if len(orderBys) == 0 {
		orderBys = append(orderBys, "created_at "+sortOrder)
	}

	query = query.OrderBy(orderBys...)

	// 分页
	offset := (list.Page - 1) * list.PageSize
	query = query.Limit(uint64(list.PageSize)).Offset(uint64(offset))

	// 组装最终 SQL
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*Todos
	err = m.conn.QueryRowsCtx(ctx, &rows, sql, args...)
	return rows, err
}
