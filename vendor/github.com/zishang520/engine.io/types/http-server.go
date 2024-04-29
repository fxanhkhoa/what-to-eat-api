package types

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/zishang520/engine.io/events"
	"golang.org/x/net/http2"
)

type HttpServer struct {
	events.EventEmitter
	*ServeMux

	servers []*http.Server
	mu      sync.RWMutex
}

func CreateServer(defaultHandler http.Handler) *HttpServer {
	s := &HttpServer{
		EventEmitter: events.New(),
		ServeMux:     NewServeMux(defaultHandler),
	}
	return s
}

func (s *HttpServer) server(addr string) *http.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	server := &http.Server{Addr: addr, Handler: s}
	server.RegisterOnShutdown(func() {
		s.Emit("close")
	})

	s.servers = append(s.servers, server)

	return server
}

func (s *HttpServer) Close(fn Callable) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.servers != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, server := range s.servers {
			if err := server.Shutdown(ctx); err != nil {
				return err
			}
		}
		if fn != nil {
			defer fn()
		}
	}
	return nil
}

func (s *HttpServer) Listen(addr string, fn Callable) *HttpServer {

	go func() {
		if err := s.server(addr).ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

func (s *HttpServer) ListenTLS(addr string, certFile string, keyFile string, fn Callable) *HttpServer {

	go func() {
		if err := s.server(addr).ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

func (s *HttpServer) ListenHTTP2TLS(addr string, certFile string, keyFile string, conf *http2.Server, fn Callable) *HttpServer {

	go func() {
		server := s.server(addr)
		if err := http2.ConfigureServer(server, conf); err != nil {
			panic(err)
		}
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}
