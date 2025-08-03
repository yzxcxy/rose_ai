package types

import "github.com/zeromicro/x/errors"

const (
	ErrorInvalidParamsCode = iota + 10001
	ErrorUserNotFound
	ErrorUserPasswordNotCorrect
)

var errorMessage = map[int]string{
	ErrorInvalidParamsCode:      "Invalid parameters",
	ErrorUserNotFound:           "User not found",
	ErrorUserPasswordNotCorrect: "User password not correct",
}

// GetError returns the error message for a given error code.
func GetError(code int) error {
	if msg, exists := errorMessage[code]; exists {
		return errors.New(code, msg)
	}
	return errors.New(code, "Unknown error")
}
