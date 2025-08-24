package logic

import (
	"context"
	"errors"
	"rose/internal/utils"
	"rose/model"
	"time"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTodoLogic {
	return &GetTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTodoLogic) GetTodo(req *types.GetTodoReq) (resp *types.Todo, err error) {
	userID, userName, err := utils.GetUserIdAndUserNameFromContext(l.ctx)
	if err != nil {
		return nil, types.GetError(types.ErrorUserNotFound)
	}
	todoModel := model.NewTodosModel(l.svcCtx.Mysql)
	todo, err := todoModel.FindOneByTodoId(l.ctx, req.TodoId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, types.GetError(types.ErrorTodoNotFound)
		}
		logx.Error("Failed to find todo:", err)
		return nil, types.GetError(types.ErrorInternalServer)
	}

	if todo.UserId != userID {
		return nil, types.GetError(types.ErrorNoPermission)
	}

	if todo.IsDeleted == 1 {
		return nil, types.GetError(types.ErrorTodoNotFound)
	}

	return &types.Todo{
		TodoId:      todo.TodoId,
		UserId:      userID,
		UserName:    userName,
		Description: todo.Description.String,
		Name:        todo.Name,
		Status:      todo.Status,
		Priority:    todo.Priority,
		DueDate:     todo.DueDate.Format(time.DateTime),
		CreatedAt:   todo.CreatedAt.Format(time.DateTime),
		UpdatedAt:   todo.UpdatedAt.Format(time.DateTime),
	}, nil
}
