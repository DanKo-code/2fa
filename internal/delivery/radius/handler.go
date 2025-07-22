package radius

import (
	"context"
	"log/slog"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"gitlab.indev.by/2fa/2fa_go_be/internal/domain/user"
	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

type Handler struct {
	authService service.Auth
}

func NewHandler(authService service.Auth) Handler {
	return Handler{
		authService: authService,
	}
}

func (h Handler) CheckSecondFactor(w radius.ResponseWriter, r *radius.Request) {
	ctx := context.Background()

	username := rfc2865.UserName_GetString(r.Packet)

	err := h.authService.SendPush(ctx, user.User{
		Email: username,
	})
	if err != nil {
		slog.Error("failed to CheckSecondFactor", "email", username, "error", err)
		if err := w.Write(r.Response(radius.CodeAccessReject)); err != nil {
			slog.Error("failed to write RADIUS reject response", "error", err)
		}
		return
	}

	if err := w.Write(r.Response(radius.CodeAccessAccept)); err != nil {
		slog.Error("failed to write RADIUS accept response", "error", err)
	}
}
