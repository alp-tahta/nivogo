package server

import (
	"context"
	"net/http"
	"time"
)

var srv *http.Server

func Init(port string, mux *http.ServeMux) error {
	srv = &http.Server{
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

func Shutdown(ctx context.Context) error {
	if srv != nil {
		return srv.Shutdown(ctx)
	}
	return nil
}
