package logic

import (
	"context"
	"database/sql"
	"errors"
	"rose/internal/utils"
	"rose/model"
	"strings"
	"time"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTodosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTodosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTodosLogic {
	return &ListTodosLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTodosLogic) ListTodos(req *types.ListTodoReq) (resp *types.ListTodoResp, err error) {

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10 // 默认每页10条
	}

	// 验证状态
	if req.Status != nil {
		if len(req.Status) > 3 {
			return nil, types.GetError(types.ErrorInvalidParamsCode)
		}
		validStatuses := map[string]struct{}{
			"pending":     {},
			"in_progress": {},
			"completed":   {},
		}
		for _, status := range req.Status {
			if _, ok := validStatuses[status]; !ok {
				return nil, types.GetError(types.ErrorInvalidParamsCode) // 无效的状态
			}
		}
	}

	// 验证优先级
	if req.Priority != nil {
		if len(req.Priority) > 3 {
			return nil, types.GetError(types.ErrorInvalidParamsCode)
		}
		validPriorities := map[string]struct{}{
			"low":    {},
			"medium": {},
			"high":   {},
		}
		for _, priority := range req.Priority {
			if _, ok := validPriorities[priority]; !ok {
				return nil, types.GetError(types.ErrorInvalidParamsCode) // 无效的优先级
			}
		}
	}

	// 验证排序字段
	if req.SortBy != nil {
		validSortFields := map[string]struct{}{
			"todo_id":     {},
			"user_id":     {},
			"description": {},
			"status":      {},
			"priority":    {},
			"due_date":    {},
			"created_at":  {},
			"updated_at":  {},
		}
		for _, sortField := range req.SortBy {
			if _, ok := validSortFields[sortField]; !ok {
				return nil, types.GetError(types.ErrorInvalidParamsCode) // 无效的排序字段
			}
		}
	}

	// 验证排序顺序
	if req.SortOrder != nil {
		if len(req.SortOrder) != len(req.SortBy) {
			return nil, types.GetError(types.ErrorInvalidParamsCode) // 排序字段和排序顺序数量不匹配
		}
		validSortFields := map[string]struct{}{
			"asc":  {},
			"desc": {},
		}

		for _, sortOrder := range req.SortOrder {
			if _, ok := validSortFields[strings.ToLower(sortOrder)]; !ok {
				return nil, types.GetError(types.ErrorInvalidParamsCode) // 无效的排序顺序
			}
		}
	}

	// 解析日期
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, err = time.Parse(time.DateTime, req.StartDate)
		if err != nil {
			return nil, types.GetError(types.ErrorInvalidDateFormat)
		}
	}

	if req.EndDate != "" {
		endDate, err = time.Parse(time.DateTime, req.EndDate)
		if err != nil {
			return nil, types.GetError(types.ErrorInvalidDateFormat)
		}
	}

	if endDate.Before(startDate) {
		return nil, types.GetError(types.ErrorInvalidDateFormat)
	}

	// 解析到期日期的范围
	var minDueDate, maxDueDate time.Time
	if req.MinDueDate != "" {
		minDueDate, err = time.Parse(time.DateTime, req.MinDueDate)
		if err != nil {
			return nil, types.GetError(types.ErrorInvalidDateFormat)
		}
	}

	if req.MaxDueDate != "" {
		maxDueDate, err = time.Parse(time.DateTime, req.MaxDueDate)
		if err != nil {
			return nil, types.GetError(types.ErrorInvalidDateFormat)
		}
	}

	if maxDueDate.Before(minDueDate) {
		return nil, types.GetError(types.ErrorInvalidDateFormat)
	}

	uid, _, err := utils.GetUserIdAndUserNameFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	todosModel := model.NewTodosModel(l.svcCtx.Mysql)
	list := model.ListToDos{
		UserId:     uid,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Status:     req.Status,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		MinDueDate: minDueDate,
		MaxDueDate: maxDueDate,
		Search:     req.Search,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	todos, err := todosModel.ListToDos(l.ctx, list)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, types.GetError(types.ErrorInternalServer)
		}
	}

	_, userName, _ := utils.GetUserIdAndUserNameFromContext(l.ctx)
	var result []types.Todo
	for _, todo := range todos {
		result = append(result, types.Todo{
			TodoId:      todo.TodoId,
			UserId:      todo.UserId,
			UserName:    userName,
			Name:        todo.Name,
			Description: todo.Description.String,
			Status:      todo.Status,
			Priority:    todo.Priority,
			DueDate:     todo.DueDate.Format(time.DateTime),
			CreatedAt:   todo.CreatedAt.Format(time.DateTime),
			UpdatedAt:   todo.UpdatedAt.Format(time.DateTime),
		})
	}

	return &types.ListTodoResp{
		List:  result,
		Total: int64(len(todos)),
	}, nil
}
