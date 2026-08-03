package portxd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

type Server struct {
	remoteAddr string
	localAddr  string
	serverPort int
	listener   net.Listener
	wg         sync.WaitGroup
	logger     *slog.Logger
}

func (s *Server) Start(ctx context.Context) error {
	if s.logger == nil {
		s.logger = slog.Default()
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.serverPort))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", s.serverPort, err)
	}
	s.listener = ln

	s.logger.Info("PortXD server started",
		"local_addr", s.localAddr,
		"server_port", s.serverPort,
	)

	s.wg.Add(1)
	go s.serve(ctx)

	return nil
}

func (s *Server) serve(ctx context.Context) {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		s.wg.Add(1)
		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, remote net.Conn) {
	defer s.wg.Done()
	defer remote.Close()

	local, err := net.Dial("tcp", s.localAddr)
	if err != nil {
		s.logger.Error("dial local", "addr", s.localAddr, "error", err)
		return
	}
	defer local.Close()

	s.logger.Debug("PortXD connection established",
		"remote", remote.RemoteAddr(),
		"local_proxy", s.localAddr,
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(local, remote)
	}()

	go func() {
		defer wg.Done()
		io.Copy(remote, local)
	}()

	wg.Wait()
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	if s.logger != nil {
		s.logger.Info("PortXD server stopped")
	}
}
