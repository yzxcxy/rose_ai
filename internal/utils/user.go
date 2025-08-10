package utils

import (
	"context"
	"encoding/json"
	"rose/internal/types"
	"strconv"
)

func GetUserIdAndUserNameFromContext(ctx context.Context) (int64, string, error) {
	// 假设 ctx 包含用户信息
	// 这里需要根据实际情况获取用户 ID 和用户名
	// 例如，如果使用 JWT 或其他认证方式，可以从 token 中解析出用户信息
	userId, ok := ctx.Value("uid").(json.Number)
	if !ok {
		err := types.GetError(types.ErrorUserNotFound)
		return 0, "", err
	}
	uid, _ := strconv.ParseInt(userId.String(), 10, 64)

	var name string
	if name, ok = ctx.Value("username").(string); !ok {
		err := types.GetError(types.ErrorUserNotFound)
		return 0, "", err
	}

	return int64(uid), name, nil
}
