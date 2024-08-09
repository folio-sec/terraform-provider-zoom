package util

import (
	"errors"

	"github.com/ogen-go/ogen/validate"
)

func IsUnexpectedStatusCodeError(err error, code int) bool {
	var unexpected *validate.UnexpectedStatusCodeError
	if errors.As(err, &unexpected) {
		return unexpected.StatusCode == code
	}
	return false
}
