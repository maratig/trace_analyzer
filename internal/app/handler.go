package app

import (
	"net/http"
	"strings"
)

const sourcePathUrlParam = "source_path"

func (a *App) RunTraceEventsListening(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(r.Method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only POST method is allowed"))
	}

	sourcePath := r.FormValue(sourcePathUrlParam)
	if sourcePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sourcePathUrlParam + " is required"))
	}

	if err := a.ListenTraceEvents(r.Context(), sourcePath); err != nil {
		// TODO return an appropriate error with appropriate HTTP code
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
