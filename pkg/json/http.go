package json

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.unanet.io/devops/go/pkg/errors"
)

func ParseBody(r *http.Request, model interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(model); err != nil {
		if err.Error() == "EOF" {
			return errors.RestError{
				Code:    400,
				Message: fmt.Sprintf("Missing POST Body"),
			}
		} else {
			return errors.RestError{
				Code:    400,
				Message: fmt.Sprintf("Invalid Post Body: %s", err),
			}
		}
	}

	if err := validation.ValidateWithContext(r.Context(), model); err != nil {
		switch err := err.(type) {
		case validation.Errors:
			return errors.RestError{
				Code:    400,
				Message: err.Error(),
			}
		default:
			return fmt.Errorf("unexpected validation error: %w", err)
		}

	}

	return nil
}
