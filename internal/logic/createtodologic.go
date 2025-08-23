package logic

import (
	"context"
	"database/sql"
	"rose/internal/utils"
	"rose/model"
	"rose/pkg/snowflake"
	"time"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTodoLogic {
	return &CreateTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTodoLogic) CreateTodo(req *types.CreateTodoReq) (resp *types.CreateTodoResp, err error) {
	// 获得用户ID
	uid, _, err := utils.GetUserIdAndUserNameFromContext(l.ctx)
	if err != nil {
		logx.Error("Failed to get user ID from context:", err)
		return nil, types.GetError(types.ErrorUserNotFound)
	}

	// dueDate 解析，如果为空，就设置为今天结束，如果不为空，就解析为time.date
	var t time.Time
	if req.DueDate == "" {
		// 设置为今天结束
		t = time.Now().Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		t, err = time.Parse("2006-01-02 15:04:05", req.DueDate)
	}
	if err != nil {
		logx.Error("Failed to parse due date:", err)
		err = types.GetError(types.ErrorInvalidDueDate)
		return
	}

	todoID := snowflake.GenID()

	todo := &model.Todos{
		UserId:      uid,
		TodoId:      todoID,
		Name:        req.Name,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     t,
		Description: sql.NullString{String: req.Description, Valid: true},
	}

	// 获得todos model
	todosModel := model.NewTodosModel(l.svcCtx.Mysql)

	// 插入待办事项
	_, err = todosModel.Insert(l.ctx, todo)
	if err != nil {
		return nil, err
	}

	return &types.CreateTodoResp{
		TodoId: todoID,
	}, nil
}
