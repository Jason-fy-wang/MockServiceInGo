package api

import (
	"net/http"
	"net/http/pprof"
)

type DebugHandler struct {
	Server http.Server
}

func (h *DebugHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (h *DebugHandler) headerCheck(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Debug-Token")

		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token != "secret-token" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *DebugHandler) Start() error {
	mux := http.NewServeMux()
	h.Register(mux)
	wrapper := h.headerCheck(mux)
	h.Server = http.Server{
		Addr:    ":6060",
		Handler: wrapper,
	}

	return h.Server.ListenAndServe()
}

func (h *DebugHandler) Stop() error {
	return h.Server.Close()
}
