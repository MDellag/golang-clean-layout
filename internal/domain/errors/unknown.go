package errors

import (
	"clean-arq-layout/internal/domain/constants"
	"fmt"
)

type UnknownError struct {
	message string
}

func NewUnknownError(message string) UnknownError {
	return UnknownError{message: message}
}

func (u UnknownError) Error() string {
	return fmt.Sprintf(constants.UnknownErrorMessage, u.message)
}
