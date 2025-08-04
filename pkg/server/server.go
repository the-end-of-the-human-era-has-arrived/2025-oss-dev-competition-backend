package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type Server struct {
	host   string
	port   string
	server *http.Server
	mux    *http.ServeMux
}

func NewServer(appCfg *AppConfig) *Server {
	return &Server{
		host: appCfg.Host,
		port: appCfg.Port,
	}
}

func (s *Server) AddRoute(handlerMap map[string]func(http.ResponseWriter, *http.Request)) {
	for path, handleFn := range handlerMap {
		s.mux.HandleFunc(path, handleFn)
	}
}

func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.host, s.port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s:%s", s.host, s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
	return nil
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mindmap-server",
		Short: "A mindmap server application",
		Long:  `A mindmap server application that manages keyword nodes and their relationships for creating mind maps.`,
		RunE:  exec,
	}

	return cmd
}

func exec(cmd *cobra.Command, args []string) error {
	var configPath string
	cmd.Flags().StringVarP(&configPath, "config", "c", "config/config.json", "Path to the configuration file")

	config := DefaultConfig()
	if err := ReadConfig(config, configPath); err != nil {
		return err
	}

	cmd.Flags().StringVarP(&config.Port, "port", "p", "8080", "Port to run the server on")
	cmd.Flags().StringVarP(&config.Host, "host", "H", "0.0.0.0", "Host to bind the server to")

	server := NewServer(config)
	// TODO: AddRoute() Handler 추가

	return server.Start()
}
