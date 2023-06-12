package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
)

func parserArgument(r *http.Request, x any) error {
	if r.Method == http.MethodGet {
		return binding.Form.Bind(r, x)
	}
	return binding.JSON.Bind(r, x)
}

func renderJSON(w http.ResponseWriter, data any) error {
	return render.JSON{Data: data}.Render(w)
}

type businessError struct {
	Code    int
	Message string
}

func (e *businessError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func Error(code int, message string) error {
	return &businessError{Code: code, Message: message}
}

func Handle[T any, K any](fn func(ctx context.Context, input T) (out K, err error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args T
		if err := parserArgument(r, &args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ret, err := fn(r.Context(), args)
		if err != nil {
			var be *businessError
			if errors.As(err, &be) {
				w.Header().Set("X-Error-Code", strconv.Itoa(be.Code))
				w.Write([]byte(be.Error()))
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-Error-Code", "0")
		renderJSON(w, ret)
	}
}
