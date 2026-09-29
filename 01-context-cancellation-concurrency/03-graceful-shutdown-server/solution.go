package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	WORKER_POOL_SIZE = 10
	CHANNEL_CAPACITY = 100
	TIMEOUT_DURATION = 10
	SERVER_PORT      = 8000
)

type Server struct {
	Ctx      context.Context
	Port     int
	Handlers *chi.Mux
	Timeout  time.Duration
	// DBConnection    string // TO BE IMPLEMENTED LATER
	WorkersPoolSize int
	WorkerPool      chan Job
}

func ValidateServerConfig(srv *Server) error {
	if srv.WorkersPoolSize <= 0 {
		return errors.New("invalid number of workers")
	}
	return nil
}

func (srv *Server) NewServer() *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.Port),
		Handler: srv.Handlers,
		// Needs to be a timeout for each
		ReadTimeout:  srv.Timeout,
		WriteTimeout: srv.Timeout,
	}
}

func (srv *Server) Start() error {
	if errValidate := ValidateServerConfig(srv); errValidate != nil {
		server := srv.NewServer()
		log.Printf("Server starting on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			return err
		}
	}

	// Initialise workers that watch channels
	for i := 0; i < srv.WorkersPoolSize; i++ {
		go worker(srv.Ctx, srv.WorkerPool)
	}

	return nil
}

func (srv *Server) Stop(ctx context.Context) {
	log.Println("Shutting down server...")
}

type Job struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Handler http.HandlerFunc
	Done    chan struct{}
}

func main() {
	srv := Server{
		Ctx:     context.Background(),
		Port:    SERVER_PORT,
		Timeout: TIMEOUT_DURATION * time.Second,
		// DBConnection: "",
		WorkersPoolSize: WORKER_POOL_SIZE,
		WorkerPool:      make(chan Job, CHANNEL_CAPACITY),
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", srv.RootEndpoint)
	srv.Handlers = r

	// Start server
	srv.Start()

	// Sleep for 1 second before simulating requests
	time.Sleep(1)
	for i := 0; i < 20; i++ {
		go srv.SimulateRequests(i)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	srv.Stop(srv.Ctx)
}

func worker(ctx context.Context, job <-chan Job) {
	for j := range job {
		j.Handler(j.Writer, j.Request)
		j.Done <- struct{}{}
	}
}

func (srv *Server) RootEndpoint(w http.ResponseWriter, r *http.Request) {
	job := Job{
		Writer:  w,
		Request: r,
		Handler: srv.HandleRoot,
		Done:    make(chan struct{}),
	}
	srv.WorkerPool <- job
	<-job.Done
}

func (srv *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	time.Sleep(500 * time.Millisecond)
	w.Write([]byte("I did something!"))
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) SimulateRequests(n int) {
	slog.Info("Running request", "request number", n+1)
	http.Get(fmt.Sprintf("http://localhost:%d/", srv.Port))
}
