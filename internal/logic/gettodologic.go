package logic

import (
	"context"
	"errors"
	"rose/model"

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
	todoModel := model.NewTodosModel(l.svcCtx.Mysql)
	todo, err := todoModel.FindOneByTodoId(l.ctx, req.TodoId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, types.GetError(types.ErrorTodoNotFound)
		}
		logx.Error("Failed to find todo:", err)
		return nil, types.GetError(types.ErrorInternalServer)
	}
	return &types.Todo{
		TodoId: todo.TodoId,
	}, nil
}
