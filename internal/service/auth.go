package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	ffrep "gitlab.indev.by/2fa/2fa_go_be/internal/domain/first_factor"
	fffarep "gitlab.indev.by/2fa/2fa_go_be/internal/domain/first_factor_failed_attempt"
	sfrep "gitlab.indev.by/2fa/2fa_go_be/internal/domain/second_factor"
	"gitlab.indev.by/2fa/2fa_go_be/internal/domain/user"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external/rocketsms"
)

type AuthProvider interface {
	GetPhone(user user.User) (string, error)
}

type NotifyProvider interface {
	SendNotify(user user.User) (string, <-chan bool, error)
}

type PushProvider interface {
	SendPush(fcmToken string, title string, body string) error
}

type Auth struct {
	ap      AuthProvider
	np      NotifyProvider
	pp      PushProvider
	fFRep   ffrep.Repository
	sFRep   sfrep.Repository
	fFFARep fffarep.Repository
	config  config.Config
	waiters *sync.Map
}

func NewAuth(
	ap AuthProvider,
	np NotifyProvider,
	pp PushProvider,
	fFRep ffrep.Repository,
	sFRep sfrep.Repository,
	fFFARep fffarep.Repository,
	config config.Config,
	waiters *sync.Map) Auth {
	return Auth{
		ap:      ap,
		np:      np,
		pp:      pp,
		fFRep:   fFRep,
		sFRep:   sFRep,
		fFFARep: fFFARep,
		config:  config,
		waiters: waiters,
	}
}

type PushResp struct {
	Status bool
	Code   string
}

func (auth Auth) SignIn(ctx context.Context, user user.User) error {
	operation := "SignIn"

	_, err := auth.sFRep.GetSecondFactorCode(ctx, user.Email)
	if err == nil {
		return apperrors.NewRecordAlreadyExists(operation, err, apperrors.ErrSecondFactorCodeAlreadyExist.Error())
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}
	_, err = auth.fFRep.GetFirstFactorCode(ctx, user.Email)
	if err == nil {
		return apperrors.NewRecordAlreadyExists(operation, err, apperrors.ErrFirstFactorCodeAlreadyExist.Error())
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	phone, err := auth.ap.GetPhone(user)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			return apperrors.NewUnauthorized(operation, err, apperrors.ErrUnauthorized.Error())
		}

		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	user.Phone = phone

	fFCode, ch, err := auth.np.SendNotify(user)
	if err != nil {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	go auth.checkingNotifySuccess(ch, user.Email)

	err = auth.fFRep.SetFirstFactorCode(ctx, user.Email, fFCode)
	if err != nil {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	return nil
}
func (auth Auth) checkingNotifySuccess(ch <-chan bool, email string) {
	operation := "checkingNotifySuccess"

	if status, ok := <-ch; ok && !status {
		err := auth.fFRep.DeleteFirstFactorCode(context.Background(), email)
		if err != nil {
			slog.Error(apperrors.Wrap(operation, err).Error())
		}
	}
}

func (auth Auth) SignInWithoutSMS(ctx context.Context, user user.User) (string, error) {
	operation := "SignInWithoutSMS"

	_, err := auth.sFRep.GetSecondFactorCode(ctx, user.Email)
	if err == nil {
		return "", apperrors.NewRecordAlreadyExists(operation, err, apperrors.ErrSecondFactorCodeAlreadyExist.Error())
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return "", apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	_, err = auth.fFRep.GetFirstFactorCode(ctx, user.Email)
	if err == nil {
		return "", apperrors.NewRecordAlreadyExists(operation, err, apperrors.ErrFirstFactorCodeAlreadyExist.Error())
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return "", apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	phone, err := auth.ap.GetPhone(user)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			return "", apperrors.NewUnauthorized(operation, err, apperrors.ErrUnauthorized.Error())
		}

		return "", apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	user.Phone = phone

	code, err := rocketsms.GenerateSMSCode()
	if err != nil {
		return "", apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	err = auth.fFRep.SetFirstFactorCode(ctx, user.Email, code)
	if err != nil {
		return "", apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	return code, nil
}

func (auth Auth) VerifyCode(ctx context.Context, user user.User) error {
	operation := "VerifyCode"

	_, err := auth.sFRep.GetSecondFactorCode(ctx, user.Email)
	if err == nil {
		return apperrors.NewRecordAlreadyExists(operation, err, apperrors.ErrSecondFactorCodeAlreadyExist.Error())
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}
	code, err := auth.fFRep.GetFirstFactorCode(ctx, user.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.NewRecordNotFound(operation, err, "first-factor code not found")
		}

		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	if code != user.FirstFactorCode {
		return apperrors.NewUnauthorized(operation, apperrors.ErrUnauthorized, "not valid first-factor code")
	}

	err = auth.fFRep.DeleteFirstFactorCode(ctx, user.Email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	err = auth.sFRep.SetSecondFactorCode(ctx, user.Email, user.SecondFactorCode)
	if err != nil {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	err = auth.fFFARep.DeleteFirstFactorFailedAttempts(ctx, user.Email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	return nil
}
func (auth Auth) checkFailedAttempts(ctx context.Context, user user.User) error {
	fa, err := auth.fFFARep.GetAndAddFirstFactorFailedAttempts(ctx, user.Email)
	if err != nil {
		return err
	}
	if int(fa) >= auth.config.RedisFirstFactorFailedAttempts {
		err := auth.fFFARep.DeleteFirstFactorFailedAttempts(ctx, user.Email)
		if err != nil {
			return err
		}
		err = auth.fFRep.DeleteFirstFactorCode(ctx, user.Email)
		if err != nil {
			return err
		}
	}
	return nil
}

func (auth Auth) SendPush(ctx context.Context, user user.User) error {
	operation := "SendPush"

	sf, err := auth.sFRep.GetSecondFactorCode(ctx, user.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.NewRecordNotFound(operation, err, apperrors.ErrNotFound.Error())
		}
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	sessionID := uuid.NewString()
	slog.Info("email:%s, sessionID:%s", user.Email, sessionID)

	err = auth.pp.SendPush(sf, auth.config.RocketSmsTitle, sessionID)
	if err != nil {
		return apperrors.NewCompositeWithContext(operation, err, apperrors.ErrUnknown.Error())
	}

	err = auth.handleAnswer(user, sessionID)
	if err != nil {
		return apperrors.Wrap(operation, err)
	}

	return nil
}
func (auth Auth) handleAnswer(user user.User, sessionID string) error {
	operation := "handleAnswer"

	ch := make(chan PushResp, 1)
	defer auth.waiters.Delete(user.Email + ":" + sessionID)

	auth.waiters.Store(user.Email+":"+sessionID, ch)

	select {
	case res := <-ch:
		if _, ok := auth.waiters.LoadAndDelete(user.Email + ":" + res.Code); !ok {
			return apperrors.Wrap(
				operation,
				apperrors.Wrap(fmt.Sprintf("filed to delete second-factor approve session: %s", sessionID),
					apperrors.ErrUnauthorized),
			)
		}
		if res.Status {
			return nil
		}
		return apperrors.Wrap(operation, apperrors.ErrPushRejected)

	case <-time.After(auth.config.PushSessionTTLInMinutes):
		return apperrors.Wrap(
			operation,
			apperrors.Wrap(fmt.Sprintf("second-factor approve session time out: %s", sessionID),
				apperrors.ErrUnauthorized),
		)
	}
}

func (auth Auth) SendPushAnswer(_ context.Context, user user.User, pushAnswerSID string, answer bool) error {
	operation := "SendPushAnswer"

	value, ok := auth.waiters.Load(user.Email + ":" + pushAnswerSID)
	if !ok {
		return apperrors.NewUnauthorized(
			operation,
			fmt.Errorf("failed get second-factor session: %s", user.Email+":"+pushAnswerSID),
			"failed get second-factor session",
		)
	}

	ch, ok := value.(chan PushResp)
	if !ok {
		return apperrors.NewUnauthorized(
			operation,
			fmt.Errorf("failed get second-factor session: %s", user.Email+":"+pushAnswerSID),
			"failed get second-factor session",
		)
	}

	ch <- PushResp{
		Status: answer,
		Code:   pushAnswerSID,
	}
	close(ch)

	return nil
}
