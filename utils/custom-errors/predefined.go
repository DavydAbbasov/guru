package customerrors

import "net/http"

// Treat as immutable; shared via errMapper.
var (
	InternalErr = &CustomErr{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeInternal,
		Message:    "internal server error",
	}

	NotFoundErr = &CustomErr{
		StatusCode: http.StatusNotFound,
		Code:       CodeNotFound,
		Message:    "not found",
	}

	AlreadyExistsErr = &CustomErr{
		StatusCode: http.StatusConflict,
		Code:       CodeAlreadyExists,
		Message:    "already exists",
	}

	ValidationErr = &CustomErr{
		StatusCode: http.StatusBadRequest,
		Code:       CodeValidation,
		Message:    "validation failed",
	}

	DatabaseRequestErr = &CustomErr{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeDatabaseRequest,
		Message:    "database request failed",
	}
)
