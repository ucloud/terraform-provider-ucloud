package umongodb

import "fmt"

const notFoundCode = "Notfound"

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: notFoundCode, message: message}
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && providerErr.errorCode == notFoundCode
}
