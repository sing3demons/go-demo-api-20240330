package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type IRouter interface {
	GET(pattern string, handler http.HandlerFunc)
	POST(pattern string, handler http.HandlerFunc)
	PUT(pattern string, handler http.HandlerFunc)
	DELETE(pattern string, handler http.HandlerFunc)
	PATCH(pattern string, handler http.HandlerFunc)
	StartHTTP(appName, port string)
}

type myRouter struct {
	*http.ServeMux
}

func NewRouter() IRouter {
	mux := http.NewServeMux()
	return &myRouter{mux}
}

func (r *myRouter) GET(pattern string, handler http.HandlerFunc) {
	r.HandleFunc("GET "+pattern, handler)
}

func (r *myRouter) POST(pattern string, handler http.HandlerFunc) {
	r.HandleFunc("POST "+pattern, handler)
}

func (r *myRouter) PUT(pattern string, handler http.HandlerFunc) {
	r.HandleFunc("PUT "+pattern, handler)
}

func (r *myRouter) DELETE(pattern string, handler http.HandlerFunc) {
	r.HandleFunc("DELETE "+pattern, handler)
}

func (r *myRouter) PATCH(pattern string, handler http.HandlerFunc) {
	r.HandleFunc("PATCH "+pattern, handler)
}

const XSession = "X-Session"

type ContextKey string

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := w.Header().Get(XSession)
		if reqId == "" {
			reqId = uuid.NewString()
			w.Header().Set(XSession, reqId)
		}
		// Set the logger in the context
		ctx := context.WithValue(r.Context(), ContextKey(XSession), reqId)
		r = r.WithContext(ctx)
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func Session(ctx context.Context) string {
	return ctx.Value(ContextKey(XSession)).(string)
}

func (r *myRouter) StartHTTP(appName, port string) {
	var (
		read  int = 10
		write int = 10
	)

	readTimeout := os.Getenv("TIMEOUT_READ")
	if readTimeout != "" {
		read, _ = strconv.Atoi(readTimeout)
	}

	writeTimeout := os.Getenv("TIMEOUT_WRITE")
	if writeTimeout != "" {
		write, _ = strconv.Atoi(writeTimeout)
	}

	var wait time.Duration

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      Logger(r),
		ReadTimeout:  time.Duration(read) * time.Second,
		WriteTimeout: time.Duration(write) * time.Second,
	}

	log.Printf("Time zone is set to %s", time.Local.String())
	log.Printf("Starting %s on port :: %s", appName, port)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("error starting server, %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}
	log.Println("server exiting")

}
