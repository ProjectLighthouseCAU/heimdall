package model

import "fmt"

// Indicates that a resource was not found
type NotFoundError struct {
	Message string
	Err     error
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
func (e NotFoundError) Unwrap() error {
	return e.Err
}
func (e NotFoundError) Status() int {
	return 404
}

// Indicates an arbitrary internal server error
type InternalServerError struct {
	Message string
	Err     error
}

func (e InternalServerError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
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
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
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
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
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
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
func (e ForbiddenError) Unwrap() error {
	return e.Err
}
func (e ForbiddenError) Status() int {
	return 403
}
