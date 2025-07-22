package rest

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	dtohttp "gitlab.indev.by/2fa/2fa_go_be/internal/domain/http"
	"gitlab.indev.by/2fa/2fa_go_be/internal/domain/user"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
	"gitlab.indev.by/2fa/2fa_go_be/internal/service"
)

type Handler struct {
	authService service.Auth
	logger      *slog.Logger
}

func NewHandler(authService service.Auth, logger *slog.Logger) Handler {
	return Handler{
		authService: authService,
		logger:      logger,
	}
}

// @Summary		Получить статус работы сервера
// @Tags			Info
// @Description	Получить статус работы сервера
// @ID				health
// @Success		200	{string}	string	"ok"
// @Router			/health [get]
func (h Handler) Health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// @Summary		Вход
// @Tags			Auth
// @Description	Вход
// @ID				sign-in
// @Accept			json
// @Produce		json
// @Param			request	body		dtohttp.SignInRequest	true	"Параметры входа"
// @Success		200		{object}	dtohttp.Response[any]
// @Failure		400		{object}	dtohttp.Response[any]
// @Failure		401		{object}	dtohttp.Response[any]
// @Failure		500		{object}	dtohttp.Response[any]
// @Router			/sign-in [post]
func (h Handler) SignIn(c *gin.Context) {
	operation := "SignIn"

	ctx := c.Request.Context()

	var req dtohttp.SignInRequest

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		respondWithErr(apperrors.NewIncorrectInputError(operation, err, apperrors.ErrBadParams.Error()), c, h.logger)
		return
	}

	err = h.authService.SignIn(ctx, user.User{Email: req.Email, Password: req.Password})
	if err != nil {
		respondWithErr(err, c, h.logger)
		return
	}

	c.JSON(http.StatusOK, dtohttp.Response[any]{
		Message: "Sent first factor code",
	})
}

// @Summary		Вход без смс
// @Tags			Auth Without SMS
// @Description	Вход без смс (В целях debug и экономии sms запросов на rocketsms)
// @ID				sign-in-without-sms
// @Accept			json
// @Produce		json
// @Param			request	body		dtohttp.SignInRequest	true	"Параметры входа"
// @Success		200		{object}	dtohttp.Response[any]
// @Failure		400		{object}	dtohttp.Response[any]
// @Failure		401		{object}	dtohttp.Response[any]
// @Failure		500		{object}	dtohttp.Response[any]
// @Router			/sign-in-without-sms [post]
func (h Handler) SignInWithOutSMS(c *gin.Context) {
	operation := "SignInWithOutSMS"

	ctx := c.Request.Context()

	var req dtohttp.SignInRequest

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		respondWithErr(apperrors.NewIncorrectInputError(operation, err, apperrors.ErrBadParams.Error()), c, h.logger)
		return
	}

	code, err := h.authService.SignInWithoutSMS(ctx, user.User{Email: req.Email, Password: req.Password})
	if err != nil {
		respondWithErr(err, c, h.logger)
		return
	}

	c.JSON(http.StatusOK, dtohttp.Response[any]{
		Data:    code,
		Message: "Sent first factor code",
	})
}

// @Summary		Проверка кода
// @Tags			Verify code
// @Description	Проверка кода
// @ID				verify-code
// @Accept			json
// @Produce		json
// @Param			request	body		dtohttp.VerifyCodeRequest	true	"Параметры входа"
// @Success		200		{object}	dtohttp.Response[any]
// @Failure		400		{object}	dtohttp.Response[any]
// @Failure		401		{object}	dtohttp.Response[any]
// @Failure		500		{object}	dtohttp.Response[any]
// @Router			/verify-code [post]
func (h Handler) VerifyCode(c *gin.Context) {
	operation := "VerifyCode"

	ctx := c.Request.Context()

	var req dtohttp.VerifyCodeRequest

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		respondWithErr(apperrors.NewIncorrectInputError(operation, err, apperrors.ErrBadParams.Error()), c, h.logger)
		return
	}

	err = h.authService.VerifyCode(ctx, user.User{
		Email:            req.Email,
		FirstFactorCode:  req.FirstFactorCode,
		SecondFactorCode: req.FCMCode,
	})
	if err != nil {
		respondWithErr(err, c, h.logger)
		return
	}

	c.JSON(http.StatusOK, dtohttp.Response[any]{
		Message: "Successfully added 2fa",
	})
}

// @Summary		Ответ на push
// @Tags			Push answer
// @Description	Ответ на push
// @ID				push-answer
// @Accept			json
// @Produce		json
// @Param			request	body		dtohttp.PushAnswerRequest	true	"Параметры входа"
// @Success		200		{object}	dtohttp.Response[any]
// @Failure		400		{object}	dtohttp.Response[any]
// @Failure		401		{object}	dtohttp.Response[any]
// @Failure		500		{object}	dtohttp.Response[any]
// @Router			/push-answer [post]
func (h Handler) PushAnswer(c *gin.Context) {
	operation := "PushAnswer"

	ctx := c.Request.Context()

	var req dtohttp.PushAnswerRequest

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		respondWithErr(apperrors.NewIncorrectInputError(operation, err, apperrors.ErrBadParams.Error()), c, h.logger)
		return
	}

	err = h.authService.SendPushAnswer(ctx, user.User{
		Email: req.Email,
	}, req.PushAnswerSID, req.Answer)
	if err != nil {
		respondWithErr(err, c, h.logger)
		return
	}

	c.JSON(http.StatusOK, dtohttp.Response[any]{
		Message: "Successfully sent push answer",
	})
}

// @Summary		Отправить push и дожидаться ответа
// @Tags			Send push
// @Description	Отправить push и дожидаться ответа (эмуляция vpn client). Для тестирования
// @ID				send-push
// @Accept			json
// @Produce		json
// @Param			request	body		dtohttp.PushRequestRequest	true	"Параметры входа"
// @Success		200		{object}	dtohttp.Response[any]
// @Failure		400		{object}	dtohttp.Response[any]
// @Failure		401		{object}	dtohttp.Response[any]
// @Failure		500		{object}	dtohttp.Response[any]
// @Router			/send-push-for-debug [post]
func (h Handler) SendPush(c *gin.Context) {
	operation := "SendPush"

	ctx := c.Request.Context()

	var req dtohttp.PushRequestRequest

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		respondWithErr(apperrors.NewIncorrectInputError(operation, err, apperrors.ErrBadParams.Error()), c, h.logger)
		return
	}

	err = h.authService.SendPush(ctx, user.User{Email: req.Email})
	if err != nil {
		respondWithErr(err, c, h.logger)
		return
	}

	c.JSON(http.StatusOK, dtohttp.Response[any]{
		Message: "CodeAccessAccept",
	})
}
