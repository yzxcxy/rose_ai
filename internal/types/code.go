package types

import "github.com/zeromicro/x/errors"

const (
	ErrorInvalidParamsCode = iota + 10001
	ErrorUserNotFound
	ErrorUserPasswordNotCorrect
	ErrorNotSupportFileType
	ErrorNotFile
	ErrorUploadFailure
	ErrorInvalidDueDate
	ErrorTodoNotFound
	ErrorInternalServer
	ErrorInvalidDateFormat
	ErrorNoPermission
	ErrorFileNotExist
)

var errorMessage = map[int]string{
	ErrorInvalidParamsCode:      "Invalid parameters",
	ErrorUserNotFound:           "User not found",
	ErrorUserPasswordNotCorrect: "User password not correct",
	ErrorNotSupportFileType:     "Not supported file type",
	ErrorNotFile:                "Not a file",
	ErrorUploadFailure:          "File upload failed",
	ErrorInvalidDueDate:         "Invalid due date",
	ErrorTodoNotFound:           "Todo not found",
	ErrorInternalServer:         "Internal server error",
	ErrorInvalidDateFormat:      "Invalid date format, expected '2006-01-02 15:04:05'",
	ErrorNoPermission:           "No permission to access this resource",
	ErrorFileNotExist:           "file not found",
}

// GetError returns the error message for a given error code.
func GetError(code int) error {
	if msg, exists := errorMessage[code]; exists {
		return errors.New(code, msg)
	}
	return errors.New(code, "Unknown error")
}
