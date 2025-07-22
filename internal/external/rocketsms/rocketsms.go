package rocketsms

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	"gitlab.indev.by/2fa/2fa_go_be/internal/domain/user"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external"
)

type Rocketsms struct {
	config config.Config
}

func NewRocketsms(config config.Config) Rocketsms {
	return Rocketsms{
		config: config,
	}
}

const (
	QUEUED    = "QUEUED"
	SENT      = "SENT"
	DELIVERED = "DELIVERED"
	FAILED    = "FAILED"
)

type smsCost struct {
	Credits int     `json:"credits"`
	Money   float64 `json:"money"`
}

type sendResponse struct {
	ID     int      `json:"id,omitempty"`
	Status string   `json:"status,omitempty"`
	Cost   *smsCost `json:"cost,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type statusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (rs Rocketsms) SendNotify(user user.User) (string, <-chan bool, error) {
	operation := "SendNotify"

	code, err := GenerateSMSCode()
	if err != nil {
		return "", nil, apperrors.Wrap(operation, err)
	}

	phone := normalizePhone(user.Phone)
	sendRes, err := rs.sendSMS(phone, code)
	if err != nil {
		return "", nil, apperrors.Wrap(operation, err)
	}

	ch := make(chan bool, 1)

	if sendRes.Status == DELIVERED {
		return code, nil, nil
	}

	go rs.awaitDelivery(ch, sendRes.ID, rs.config.RocketSmsTimeoutAmount, rs.config.RocketSmsTimeout)

	if sendRes.Status != FAILED {
		return code, ch, nil
	}

	return "", nil, apperrors.Wrap(operation, apperrors.ErrFailedSentFirstFactor)
}
func (rs Rocketsms) sendSMS(phone, text string) (*sendResponse, error) {
	operation := "sendSMS"

	values := url.Values{}
	values.Set("username", rs.config.RocketSmsLogin)
	values.Set("password", rs.config.RocketSmsPass)
	values.Set("phone", phone)
	values.Set("text", text)
	values.Set("sender", rs.config.RocketSmsSender)
	values.Set("priority", "true")

	reqURL := "https://api.rocketsms.by/simple/send?" + values.Encode()

	resp, err := http.Post(reqURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, apperrors.Wrap(operation, apperrors.Wrap("failed to send SMS request", err))
	}
	defer external.SafeCloseBody(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(operation, apperrors.Wrap("failed to read SMS response body", err))
	}

	var res sendResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, apperrors.Wrap(operation, apperrors.Wrap("failed to parse send response", err))
	}

	if res.Error != "" {
		return nil, apperrors.Wrap(operation, apperrors.Wrap("rocketSMS error", err))
	}

	return &res, nil
}
func (rs Rocketsms) awaitDelivery(ch chan<- bool, smsID int, retries int, interval time.Duration) {
	operation := "awaitDelivery"

	defer close(ch)

	for i := 0; i < retries; i++ {
		time.Sleep(interval)

		status, err := rs.checkSMSStatus(smsID)
		if err != nil {
			ch <- false
			return
		}

		switch status {
		case DELIVERED:
			ch <- true
			return
		case FAILED:
			slog.Error(apperrors.Wrap(operation, fmt.Errorf("SMS delivery failed")).Error())
			ch <- false
			return
		}
	}

	slog.Error(apperrors.Wrap(operation, fmt.Errorf("SMS delivery timed out")).Error())
	ch <- false
}

func (rs Rocketsms) checkSMSStatus(smsID int) (string, error) {
	operation := "checkSMSStatus"

	statusURL := "https://api.rocketsms.by/simple/status"
	values := url.Values{}
	values.Set("username", rs.config.RocketSmsLogin)
	values.Set("password", rs.config.RocketSmsPass)
	values.Set("id", strconv.Itoa(smsID))
	reqURL := statusURL + "?" + values.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", apperrors.Wrap(operation, apperrors.Wrap("failed to check SMS status", err))
	}
	defer external.SafeCloseBody(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.Wrap(operation, apperrors.Wrap("failed to read SMS response body", err))
	}

	var res statusResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", apperrors.Wrap(operation, apperrors.Wrap("failed to parse status response", err))
	}

	return res.Status, nil
}

func normalizePhone(phone string) string {
	if len(phone) > 0 && phone[0] == '+' {
		return phone[1:]
	}
	return phone
}

func GenerateSMSCode() (string, error) {
	operation := "GenerateSMSCode"

	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", apperrors.Wrap(operation, apperrors.Wrap("failed to generate code", err))
	}
	code := 100000 + n.Int64()
	return fmt.Sprintf("%06d", code), nil
}
