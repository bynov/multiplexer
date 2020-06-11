package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

type InvalidURLLengthError int

func (e InvalidURLLengthError) Error() string {
	return fmt.Sprintf("invalid url len. Allowed: 0 < len < 20, got: %d", e)
}

type apiError struct {
	Message string `json:"message"`
	Code    int    `json:"status_code"`
}

func (e apiError) Error() string {
	return e.Message
}

func (e apiError) StatusCode() int {
	return e.Code
}

// getHTTPCode is used to determine status code based on error type.
// Our error need to implement StatusCoder interface so
// DefaultErrorEncoder could attach needed status code by itself
func getHTTPCode(err error) int {
	if err == io.EOF {
		return http.StatusBadRequest
	}

	switch e := err.(type) {
	case *json.SyntaxError:
		return http.StatusBadRequest
	case InvalidURLLengthError:
		return http.StatusBadRequest
	case net.Error:
		if e.Timeout() {
			return http.StatusRequestTimeout
		}

		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func errorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	err = apiError{
		Code:    getHTTPCode(err),
		Message: err.Error(),
	}

	httptransport.DefaultErrorEncoder(ctx, err, w)
}
