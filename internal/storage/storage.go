package storage

import "errors"

var (
	ErrPostNotFound = errors.New("Пост не найден")
	ErrUserNotFound = errors.New("User not found")
	ErrAppNotFound  = errors.New("App not found")
)
