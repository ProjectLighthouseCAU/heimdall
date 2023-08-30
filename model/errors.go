package model

// Indicates that a resource was not found
type NotFoundError struct {
	Message string
	Err     error
}

func (e NotFoundError) Error() string {
	s := "Not Found"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e NotFoundError) Unwrap() error {
	return e.Err
}
func (e NotFoundError) Status() int {
	return 404
}

type BadRequestError struct {
	Message string
	Err     error
}

func (e BadRequestError) Error() string {
	s := "Bad Request"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e BadRequestError) Unwrap() error {
	return e.Err
}
func (e BadRequestError) Status() int {
	return 400
}

// Indicates an arbitrary internal server error
type InternalServerError struct {
	Message string
	Err     error
}

func (e InternalServerError) Error() string {
	s := "Internal Server Error"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e InternalServerError) Unwrap() error {
	return e.Err
}
func (e InternalServerError) Status() int {
	return 500
}

// Indicates that the action is not possible
type ConflictError struct {
	Message string
	Err     error
}

func (e ConflictError) Error() string {
	s := "Conflict"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e ConflictError) Unwrap() error {
	return e.Err
}
func (e ConflictError) Status() int {
	return 409
}

// Indicates that the user is not authenticated
type UnauthorizedError struct {
	Message string
	Err     error
}

func (e UnauthorizedError) Error() string {
	s := "Unauthorized"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e UnauthorizedError) Unwrap() error {
	return e.Err
}
func (e UnauthorizedError) Status() int {
	return 401
}

// Indicates that the action is not allowed
type ForbiddenError struct {
	Message string
	Err     error
}

func (e ForbiddenError) Error() string {
	s := "Forbidden"
	if e.Message != "" {
		s = s + ": " + e.Message
	}
	if e.Err != nil {
		s = s + ": " + e.Err.Error()
	}
	return s
}
func (e ForbiddenError) Unwrap() error {
	return e.Err
}
func (e ForbiddenError) Status() int {
	return 403
}
