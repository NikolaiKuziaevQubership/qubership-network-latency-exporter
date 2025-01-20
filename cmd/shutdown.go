package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type shutdown struct {
	srv     *http.Server
	logger  log.Logger
	ctx     context.Context
	timeout time.Duration
}

func (s *shutdown) listen() {
	_ = level.Debug(s.logger).Log("msg", "Start listening for interruption signal")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	s.do()
}

func (s *shutdown) do() {
	_ = level.Info(s.logger).Log("msg", "Try to shut down server gracefully", "timeout", s.timeout)
	shtdwnCtx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	if err := s.srv.Shutdown(shtdwnCtx); err != nil {
		_ = level.Info(s.logger).Log("msg", "Failed to shut down server gracefully", "timeout", s.timeout, "err", err)
		_ = level.Info(s.logger).Log("msg", "Force closing server", "err", s.srv.Close())
	}
}
