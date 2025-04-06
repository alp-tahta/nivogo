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
		ReadTimeout:       20 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       20 * time.Second,
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
