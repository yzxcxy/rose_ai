package logic

import (
	"context"
	"database/sql"
	"rose/model"
	"time"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTodoLogic {
	return &UpdateTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTodoLogic) UpdateTodo(req *types.UpdateTodoReq) (resp *types.UpdateTodoResp, err error) {
	todoModel := model.NewTodosModel(l.svcCtx.Mysql)

	var description sql.NullString
	if req.Description == "" {
		description = sql.NullString{Valid: false} // 如果描述为空，则设置为无效
	} else {
		description = sql.NullString{String: req.Description, Valid: true} // 否则设置为有效
	}

	dueTime, err := time.Parse("2006-01-02 15:04:05", req.DueDate)
	if err != nil {
		return nil, types.GetError(types.ErrorInvalidDateFormat)
	}

	err = todoModel.Update(l.ctx, &model.Todos{
		TodoId:      req.TodoId,
		Name:        req.Name,
		Description: description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     dueTime,
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateTodoResp{}, nil
}
