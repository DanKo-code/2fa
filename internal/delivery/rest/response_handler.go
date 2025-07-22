package rest

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	dtohttp "gitlab.indev.by/2fa/2fa_go_be/internal/domain/http"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
)

func respondWithErr(err error, c *gin.Context, logger *slog.Logger) {
	var appErr apperrors.CompositeError
	ok := errors.As(err, &appErr)
	if !ok {
		c.JSON(http.StatusInternalServerError, dtohttp.NewErrorResponse(err.Error()))
		logger.Error(err.Error())
		return
	}
	switch appErr.ErrorType() {
	case apperrors.ErrorTypeIncorrectInput:
		c.JSON(http.StatusBadRequest, dtohttp.NewErrorResponse(appErr.External()))
	case apperrors.ErrorTypeRecordNotFound:
		c.JSON(http.StatusNotFound, dtohttp.NewErrorResponse(appErr.External()))
	case apperrors.ErrorTypeRecordAlreadyExists:
		c.JSON(http.StatusConflict, dtohttp.NewErrorResponse(appErr.External()))
	case apperrors.ErrorTypeUnauthorized:
		c.JSON(http.StatusUnauthorized, dtohttp.NewErrorResponse(appErr.External()))
	default:
		c.JSON(http.StatusInternalServerError, dtohttp.NewErrorResponse(appErr.External()))
	}
	logger.Error(appErr.Error())
}
