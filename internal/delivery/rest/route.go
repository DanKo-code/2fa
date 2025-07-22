package rest

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

func RegisterRoutes(router *gin.Engine, authService service.Auth, logger *slog.Logger) {
	h := NewHandler(authService, logger)

	api := router.Group("/api/2fa")
	api.GET("health", h.Health)
	api.POST("/sign-in", h.SignIn)
	api.POST("/sign-in-without-sms", h.SignInWithOutSMS)
	api.POST("/verify-code", h.VerifyCode)
	api.POST("/push-answer", h.PushAnswer)
	api.POST("/send-push-for-debug", h.SendPush)
}
