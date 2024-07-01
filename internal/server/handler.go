package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/app"
)

const (
	sourcePathUrlParam = "source_path"
	procIDParam        = "id"
)

type Handler struct {
	ctx context.Context
	app *app.App
}

func NewHandler(ctx context.Context, app *app.App) (*Handler, error) {
	if ctx == nil {
		return nil, apiError.ErrNilContext
	}
	if app == nil {
		return nil, errors.New("app must not be nil")
	}

	return &Handler{ctx: ctx, app: app}, nil
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

	if id, err := h.app.ProcessTraceSource(h.ctx, sourcePath); err != nil {
		// TODO return an appropriate error with appropriate HTTP code
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf(`{"id": %d}`, id)
		w.Write([]byte(msg))
	}
}

func (h *Handler) TopIdlingGoroutines(w http.ResponseWriter, r *http.Request) {
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

	top, err := h.app.TopIdlingGoroutines(h.ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id"))
		return
	}

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(top)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json creation error; " + err.Error()))
	}

	w.Write(data)
}
