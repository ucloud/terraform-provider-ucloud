package ucloud

import (
	"fmt"
	"strings"
)

const (
	NotFound = "Notfound"
)

type ProviderError struct {
	errorCode string
	message   string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", e.errorCode, e.message)
}

func (err *ProviderError) ErrorCode() string {
	return err.errorCode
}

func (err *ProviderError) Message() string {
	return err.message
}

func newNotFoundError(str string) error {
	return &ProviderError{
		errorCode: NotFound,
		message:   str,
	}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	if e, ok := err.(*ProviderError); ok &&
		(e.ErrorCode() == NotFound || strings.Contains(strings.ToLower(e.Message()), NotFound)) {
		return true
	}

	return false
}
