package models

import "fmt"

// Custom error types
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

type InsufficientResourcesError struct {
	Resource  string
	Required  int
	Available int
}

func (e *InsufficientResourcesError) Error() string {
	return fmt.Sprintf("insufficient %s: required %d, available %d", e.Resource, e.Required, e.Available)
}

// Helper functions for creating errors
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

func NewNotFoundError(resource, id string) error {
	return &NotFoundError{Resource: resource, ID: id}
}

func NewConflictError(message string) error {
	return &ConflictError{Message: message}
}

func NewInsufficientResourcesError(resource string, required, available int) error {
	return &InsufficientResourcesError{
		Resource:  resource,
		Required:  required,
		Available: available,
	}
}
