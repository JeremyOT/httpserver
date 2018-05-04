package httpserver

import (
	"context"
	"log"
	"net"
	"net/http"
)

// Server helps reduce boilerplate when writing tools that center around
// an http.Server instance.
type Server struct {
	quit            chan struct{}
	wait            chan struct{}
	started         chan struct{}
	handlerFunc     http.HandlerFunc
	address         net.Addr
	server          *http.Server
	shutdownHandler func()
}

// New returns a server with the specified handler.
func New(handlerFunc http.HandlerFunc) *Server {
	return &Server{
		handlerFunc: handlerFunc,
	}
}

// SetShutdownHandler lets you add a function to the shutdown pipeline. It
// will be called after http.Server.Shutdown and will block Stop and Wait until
// it returns.
func (s *Server) SetShutdownHandler(shutdownHandler func()) {
	s.shutdownHandler = shutdownHandler
}

// Address returns the server's current address.
func (s *Server) Address() net.Addr {
	return s.address
}

func (s *Server) run(listener net.Listener) {
	defer close(s.wait)
	defer func() {
		if s.shutdownHandler != nil {
			s.shutdownHandler()
		}
	}()
	s.server = &http.Server{Handler: http.HandlerFunc(s.handlerFunc)}
	defer s.server.Shutdown(context.Background())
	go s.server.Serve(listener)
	log.Println("Listening for requests on", s.Address())
	close(s.started)
	<-s.quit
}

// Start starts the Server listening on the specified address. If no port is
// specified, the Server will pick one. Use Address() after start to see which
// port was selected.
func (s *Server) Start(address string) (err error) {
	s.quit = make(chan struct{})
	s.wait = make(chan struct{})
	s.started = make(chan struct{})
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.address = listener.Addr()
	go s.run(listener)
	return nil
}

// WaitForStart returns a channel that is closed when the Server has finished
// starting and is listening for connections on Address().
func (s *Server) WaitForStart() <-chan struct{} {
	return s.started
}

// Wait returns a channel that is closed after the Server has shut down.
func (s *Server) Wait() <-chan struct{} {
	return s.wait
}

// Stop gracefully shuts down the Server and returns the channel from Wait.
// Note that it has the same limitations as http.Server.Shutdown.
func (s *Server) Stop() <-chan struct{} {
	close(s.quit)
	return s.Wait()
}