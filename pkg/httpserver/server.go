package httpserver

import (
	"context"
	"net/http"
)

type Config struct {
	Host string `env:"SERVER_HOST" env-default:"localhost"`
	Port string `env:"SERVER_PORT" env-default:"80"`
}

type Server struct {
	server *http.Server
}

func New(handler http.Handler, cfg Config) *Server {
	return &Server{
		&http.Server{
			Addr:    cfg.Host + ":" + cfg.Port,
			Handler: handler,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	return s.server.Shutdown(context.Background())
}
