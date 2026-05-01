package customerrors

import "errors"

// Keyed by Code (not error identity) so wrapped sentinels still resolve.
var errMapper = map[string]*CustomErr{
	CodeInternal:        InternalErr,
	CodeNotFound:        NotFoundErr,
	CodeAlreadyExists:   AlreadyExistsErr,
	CodeValidation:      ValidationErr,
	CodeDatabaseRequest: DatabaseRequestErr,
}

func Resolve(err error) *CustomErr {
	if err == nil {
		return InternalErr
	}
	var serviceErr Error
	if errors.As(err, &serviceErr) {
		return serviceErr.ToCustomError()
	}
	return InternalErr
}
