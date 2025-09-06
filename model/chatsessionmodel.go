package model

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"reflect"
)

var _ ChatSessionModel = (*customChatSessionModel)(nil)

type (
	// ChatSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatSessionModel.
	ChatSessionModel interface {
		chatSessionModel
		withSession(session sqlx.Session) ChatSessionModel
		UpdateIgnoreNull(ctx context.Context, data *ChatSession) error
		FindChatSessionByUserId(ctx context.Context, userId int64) ([]*ChatSession, error)
	}

	customChatSessionModel struct {
		*defaultChatSessionModel
	}
)

// NewChatSessionModel returns a model for the database table.
func NewChatSessionModel(conn sqlx.SqlConn) ChatSessionModel {
	return &customChatSessionModel{
		defaultChatSessionModel: newChatSessionModel(conn),
	}
}

func (m *customChatSessionModel) withSession(session sqlx.Session) ChatSessionModel {
	return NewChatSessionModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultChatSessionModel) FindChatSessionByUserId(ctx context.Context, userId int64) ([]*ChatSession, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `is_deleted` != 1", chatSessionRows, m.table)
	var resp []*ChatSession
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultChatSessionModel) UpdateIgnoreNull(ctx context.Context, data *ChatSession) error {
	update := sq.Update("chat_session").Where(sq.Eq{"session_id": data.SessionId})

	v := reflect.ValueOf(data).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Id" || t.Field(i).Name == "session_id" {
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
