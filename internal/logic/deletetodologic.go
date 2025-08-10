package logic

import (
	"context"
	"rose/model"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTodoLogic {
	return &DeleteTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTodoLogic) DeleteTodo(req *types.DeleteTodoReq) error {
	todosModel := model.NewTodosModel(l.svcCtx.Mysql)

	// 逻辑删除待办事项
	err := todosModel.Update(l.ctx, &model.Todos{
		TodoId:    req.TodoId,
		IsDeleted: 1,
	})
	if err != nil {
		return err
	}

	return nil
}
