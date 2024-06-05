package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/maratig/trace_analyzer/pkg/app"
)

const sourcePathUrlParam = "source_path"

type Handler struct {
	app *app.App
}

func NewHandler(app *app.App) (*Handler, error) {
	if app == nil {
		return nil, errors.New("app must not be nil")
	}

	return &Handler{app: app}, nil
}

func (h *Handler) RunTraceEventsListening(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(r.Method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only POST method is allowed"))
	}

	sourcePath := r.FormValue(sourcePathUrlParam)
	if sourcePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sourcePathUrlParam + " is required"))
	}

	if err := h.app.ListenTraceEvents(r.Context(), sourcePath); err != nil {
		// TODO return an appropriate error with appropriate HTTP code
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
