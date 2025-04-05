package server

import (
	"net/http"
	"time"
)

func Init(port string, mux *http.ServeMux) error {
	srv := http.Server{
		Addr:              port,
		Handler:           mux,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	err := srv.ListenAndServe()
	return err
}
