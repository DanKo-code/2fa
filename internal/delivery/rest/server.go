package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gitlab.indev.by/2fa/2fa_go_be/docs"
	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	"gitlab.indev.by/2fa/2fa_go_be/internal/logger"
	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

type Server struct {
	srv            *http.Server
	isShuttingDown atomic.Bool
}

func NewServer(config config.Config, logger *logger.Logger, appServices service.Auth) *Server {
	var server Server

	engine := gin.New()
	engine.Use(sloggin.New(logger.Logger))
	engine.Use(gin.Recovery())
	engine.Use(server.shuttingDownMiddleware())

	engine.GET("/api/2fa/swagger/*any", func(context *gin.Context) {
		docs.SwaggerInfo.Host = context.Request.Host
		docs.SwaggerInfo.BasePath = "/api/2fa"
		ginSwagger.CustomWrapHandler(&ginSwagger.Config{URL: "/api/2fa/swagger/doc.json"}, swaggerFiles.Handler)(context)
	})

	RegisterRoutes(engine, appServices, logger.Logger)

	srv := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              ":" + config.AppPort,
		Handler:           engine.Handler(),
	}
	server.srv = srv

	return &server
}

func (s *Server) Serve(ctx context.Context, c config.Config) error {
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	defer stopOngoingGracefully()

	s.srv.BaseContext = func(_ net.Listener) context.Context {
		return ongoingCtx
	}

	errCh := make(chan error, 1)

	go func() {
		slog.Info(fmt.Sprintf("Listening and serving HTTP on %s", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server error: %w", err)
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Received shutdown signal, shutting down HTTP server.")

		time.Sleep(c.ReadinessDrainDelay)
		slog.Info("Readiness check propagated, now waiting for ongoing requests to finish.")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), c.ShutdownPeriod)
		defer cancel()

		err := s.srv.Shutdown(shutdownCtx)

		if err != nil {
			slog.Error("Failed to wait for ongoing requests to finish. Waiting for forced shutdown.", "err", err)
			time.Sleep(c.ShutdownHardPeriod)
			return err
		}

		slog.Info("HTTP server shut down cleanly.")
		return nil

	case err := <-errCh:
		if err != nil {
			slog.Error("HTTP server exited with error", "err", err)
			return err
		}
		slog.Info("HTTP server exited cleanly.")
		return nil
	}
}

func (s *Server) shuttingDownMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.isShuttingDown.Load() {
			c.String(http.StatusServiceUnavailable, "Shutting down")

			return
		}
		c.Next()
	}
}

func CORSMiddleware(conf config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: conf.AllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PUT", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept",
			"origin", "Cache-Control", "X-Requested-With",
			"Access-Control-Allow-Headers", "Content-Type",
			"access-control-allow-origin", "access-control-allow-headers",
		},
		MaxAge: 12 * time.Hour,
	})
}
