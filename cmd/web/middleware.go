package main

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func timeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(next, timeout, http.StatusText(http.StatusServiceUnavailable))
}
