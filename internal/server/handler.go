package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/maratig/trace_analyzer/pkg/app"
)

const (
	sourcePathUrlParam = "source_path"
	procIDParam        = "id"
)

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
		return
	}

	sourcePath := r.FormValue(sourcePathUrlParam)
	if sourcePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sourcePathUrlParam + " is required"))
		return
	}

	if id, err := h.app.ListenTraceEvents(r.Context(), sourcePath); err != nil {
		// TODO return an appropriate error with appropriate HTTP code
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf(`{"id": %d}`, id)
		w.Write([]byte(msg))
	}
}

func (h *Handler) TraceEventsStat(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(r.Method, "GET") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only GET method is allowed"))
		return
	}

	idStr := r.URL.Query().Get(procIDParam)
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id is required"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id"))
		return
	}

	data, err := h.app.Stats(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id"))
		return
	}

	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf(`{"goroutines": %d}`, data)
	w.Write([]byte(msg))
}
