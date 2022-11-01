package apperror

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"learn/pkg/logging"
	"net/http"
)

type appHandler func(w http.ResponseWriter, r *http.Request, params httprouter.Params) error

func Middleware(h appHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var appErr *AppError
		err := h(w, r, p)
		logger := logging.GetLogger()
		logger.Info(err)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(ErrNotFound.Marshal())
					return
				}
				err = err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(ErrNotFound.Marshal())
				return
			}
			w.WriteHeader(http.StatusTeapot)
			w.Write(systemError(err).Marshal())
			return
		}

	}
}
