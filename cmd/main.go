package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	fblib "firebase.google.com/go/v4"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	deliveryradius "gitlab.indev.by/2fa/2fa_go_be/internal/delivery/radius"
	deliveryrest "gitlab.indev.by/2fa/2fa_go_be/internal/delivery/rest"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external/firebase"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external/keycloak"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external/rocketsms"
	"gitlab.indev.by/2fa/2fa_go_be/internal/logger"
	"gitlab.indev.by/2fa/2fa_go_be/internal/repository"
	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

//	@title			2FA
//	@version		0.1.0
//	@description	service for 2fa
//	@in				header

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf := config.LoadConfig()

	globalLogger := logger.NewLogger(conf)
	slog.SetDefault(globalLogger.Logger)

	kc := keycloak.NewKeyCloak(conf)
	rs := rocketsms.NewRocketsms(conf)
	app, err := fblib.NewApp(ctx, nil, option.WithCredentialsFile(conf.FireBasePrivateKeyPath))
	if err != nil {
		slog.Error("server init error:", err)
		os.Exit(1)
	}
	fb := firebase.NewFirebase(app)
	db, err := repository.Connect(context.Background(), conf)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fFRep := repository.NewFirstFactor(db, conf)
	sFRep := repository.NewSecondFactor(db, conf)
	fFFARep := repository.NewFirstFactorFailedAttempt(db, conf)
	waiters := sync.Map{}
	authService := service.NewAuth(
		kc,
		rs,
		fb,
		fFRep,
		sFRep,
		fFFARep,
		conf,
		&waiters)

	restSrv := deliveryrest.NewServer(conf, globalLogger, authService)
	radiusSrv := deliveryradius.NewServer(conf, globalLogger, authService)

	var g errgroup.Group

	g.Go(func() error {
		return restSrv.Serve(ctx, conf)
	})
	g.Go(func() error {
		return radiusSrv.Serve(ctx, conf)
	})

	if err := g.Wait(); err != nil {
		slog.Error("server exited with error:", err)
	} else {
		slog.Error("servers exited cleanly")
	}
}
