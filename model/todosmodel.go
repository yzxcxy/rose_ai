package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"reflect"
	"rose/internal/types"
	"strings"
	"time"
)

var _ TodosModel = (*customTodosModel)(nil)

type (
	// TodosModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTodosModel.
	TodosModel interface {
		todosModel
		withSession(session sqlx.Session) TodosModel
		ListToDos(ctx context.Context, list ListToDos) ([]*Todos, error)
		UpdateIgnoreNull(ctx context.Context, data *Todos) error
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
	//return NewTodosModel(sqlx.NewSqlConnFromSession(session))
	//为了解决版本依赖问题，临时使用下面的方式
	return &customTodosModel{}
}

func (m *customTodosModel) ListToDos(ctx context.Context, list ListToDos) ([]*Todos, error) {
	// 动态拼接sql
	if list.UserId <= 0 {
		return nil, types.GetError(types.ErrorInvalidParamsCode)
	}

	// 基础查询
	query := sq.Select(
		"id", "todo_id", "user_id", "name", "description",
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
	var sortOrder = make([]string, len(list.SortBy))
	for i := range sortOrder {
		sortOrder[i] = "DESC"
	}

	if list.SortOrder != nil {
		for i := range list.SortOrder {
			sortOrder[i] = strings.ToUpper(list.SortOrder[i])
		}
	}

	// 2. 允许的字段白名单
	allowed := map[string]struct{}{
		"id": {}, "todo_id": {}, "name": {}, "status": {}, "priority": {},
		"due_date": {}, "created_at": {}, "updated_at": {},
	}

	// 3. 遍历 SortBy 切片，过滤并拼接
	var orderBys []string
	for i, col := range list.SortBy {
		col = strings.ToLower(col)
		if _, ok := allowed[col]; ok {
			orderBys = append(orderBys, fmt.Sprintf("%s %s", col, sortOrder[i]))
		}
	}

	query = query.OrderBy(orderBys...)

	// 分页
	offset := (list.Page - 1) * list.PageSize
	query = query.Limit(uint64(list.PageSize)).Offset(uint64(offset))

	// 组装最终 SQL
	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*Todos
	err = m.conn.QueryRowsCtx(ctx, &rows, toSql, args...)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNotFound) {
			return nil, types.GetError(types.ErrorInternalServer)
		}
	}
	return rows, nil
}

func (m *defaultTodosModel) UpdateIgnoreNull(ctx context.Context, data *Todos) error {
	update := sq.Update("todos").Where(sq.Eq{"todo_id": data.TodoId})

	v := reflect.ValueOf(data).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Id" || t.Field(i).Name == "TodoId" {
			continue
		}

		dbTag := t.Field(i).Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		fv := v.Field(i)

		// 判断“空值”
		if IsZero(fv) {
			continue
		}

		update = update.Set(dbTag, fv.Interface())
	}

	query, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build sql failed: %w", err)
	}
	_, err = m.conn.Exec(query, args...)
	return err
}

// IsZero 判断值是否是“空值”
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Struct:
		// sql.NullString / time.Time 等结构体
		if v.Type() == reflect.TypeOf(sql.NullString{}) {
			return !v.Interface().(sql.NullString).Valid
		}
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		if v.Type() == reflect.TypeOf(sql.NullInt64{}) {
			return !v.Interface().(sql.NullInt64).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullFloat64{}) {
			return !v.Interface().(sql.NullFloat64).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullBool{}) {
			return !v.Interface().(sql.NullBool).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullInt32{}) {
			return !v.Interface().(sql.NullInt32).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullInt16{}) {
			return !v.Interface().(sql.NullInt16).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullByte{}) {
			return !v.Interface().(sql.NullByte).Valid
		}
		if v.Type() == reflect.TypeOf(sql.NullTime{}) {
			return !v.Interface().(sql.NullTime).Valid
		}
		// 其他结构体：直接比较零值
		return v.IsZero()
	default:
		return v.IsZero()
	}
}
