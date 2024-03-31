package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type IRouter interface {
	GET(pattern string, handler HandlerFunc)
	POST(pattern string, handler HandlerFunc)
	PUT(pattern string, handler HandlerFunc)
	DELETE(pattern string, handler HandlerFunc)
	PATCH(pattern string, handler HandlerFunc)
	StartHTTP(appName, port string)
}

type myRouter struct {
	*http.ServeMux
}

func NewRouter() IRouter {
	mux := http.NewServeMux()
	// mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte("OK"))
	// })
	return &myRouter{mux}
}

func (r *myRouter) setParams(pattern string, req *http.Request) *http.Request {
	param := `\{([^/]+)\}`
	re := regexp.MustCompile(param)
	matches := re.FindAllStringSubmatch(pattern, -1)
	if matches == nil {
		return req
	}
	p := strings.Split(pattern, "/")
	u := strings.Split(req.URL.Path, "/")
	paramMap := map[string]string{}
	for i := 0; i < len(p); i++ {
		if p[i] != "" && u[i] != "" {
			paramMap[p[i]] = u[i]
		}

	}

	for i, match := range matches {
		if paramMap[match[i]] != "" {
			req = req.WithContext(context.WithValue(req.Context(), ContextKey(match[i]), paramMap[match[i]]))
		}
	}
	return req
}

func (r *myRouter) GET(pattern string, handler HandlerFunc) {
	r.HandleFunc("GET "+pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, r.setParams(pattern, req))
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (r *myRouter) POST(pattern string, handler HandlerFunc) {
	r.HandleFunc("POST "+pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, r.setParams(pattern, req))
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (r *myRouter) PUT(pattern string, handler HandlerFunc) {
	r.HandleFunc("PUT "+pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (r *myRouter) DELETE(pattern string, handler HandlerFunc) {
	r.HandleFunc("DELETE "+pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, r.setParams(pattern, req))
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (r *myRouter) PATCH(pattern string, handler HandlerFunc) {
	r.HandleFunc("PATCH "+pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, r.setParams(pattern, req))
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

const XSession = "X-Session"

type ContextKey string

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
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

		log.Printf("%s %s %s %s %s\n", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start).String(), reqId)
	})
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

	r.GET("/healthz", func(ctx IContext) error {
		return ctx.Status(http.StatusOK)
	})

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
