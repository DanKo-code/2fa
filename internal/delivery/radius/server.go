package radius

import (
	"context"
	"log/slog"

	"layeh.com/radius"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	"gitlab.indev.by/2fa/2fa_go_be/internal/logger"
	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

type Server struct {
	srv *radius.PacketServer
}

func NewServer(config config.Config, logger *logger.Logger, appServices service.Auth) *Server {
	var server Server

	handler := NewHandler(appServices)

	server.srv = &radius.PacketServer{
		Handler:      radius.HandlerFunc(handler.CheckSecondFactor),
		SecretSource: radius.StaticSecretSource([]byte(config.RadiusSecret)),
	}

	return &server
}

func (s *Server) Serve(ctx context.Context, c config.Config) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("RADIUS server started on :1812")
		if err := s.srv.ListenAndServe(); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Received shutdown signal, shutting down RADIUS server.")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), c.ShutdownPeriod)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Graceful RADIUS shutdown failed", "err", err)
			return err
		}

		slog.Info("RADIUS server shut down.")
		return nil

	case err := <-errCh:
		if err != nil {
			slog.Error("RADIUS server exited with error", "err", err)
			return err
		}
		slog.Info("RADIUS server exited cleanly")
		return nil
	}
}
