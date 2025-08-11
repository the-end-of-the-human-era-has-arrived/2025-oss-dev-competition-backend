package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.ServerConfig) (*Server, error) {
	readTimeout, err := time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return nil, err
	}

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      newAPIServeMux(),
	}

	server := &Server{
		httpServer: httpSrv,
	}

	return server, nil
}

func (s *Server) Start() error {
	go func() {
		log.Printf("Starting server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server 30 seconds to finish handling requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
	return nil
}

func (s *Server) InstallAPIGroup(apigroups ...APIGroup) {
	mux := s.httpServer.Handler.(*APIServeMux)

	for _, group := range apigroups {
		apis := group.ListAPIs()
		mux.RegistAPI(apis...)
	}
}
