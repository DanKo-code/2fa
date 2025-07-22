package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort             string
	AllowOrigins        []string
	LogLevel            int
	ShutdownPeriod      time.Duration
	ShutdownHardPeriod  time.Duration
	ReadinessDrainDelay time.Duration

	RadiusSecret string

	KeycloakClientSecret string

	RocketSmsLogin         string
	RocketSmsPass          string
	RocketSmsSender        string
	RocketSmsTitle         string
	RocketSmsTimeout       time.Duration
	RocketSmsTimeoutAmount int

	RedisPort                      string
	RedisPass                      string
	RedisFirstFactorCodeTTLMinutes time.Duration
	RedisFirstFactorFailedAttempts int
	RedisSecondFactorCodeTTLDays   time.Duration

	FireBasePrivateKeyPath string

	PushSessionTTLInMinutes time.Duration
}

func LoadConfig(filenames ...string) Config {
	if err := godotenv.Load(filenames...); err != nil {
		panic(fmt.Sprintf("Error loading %v file", filenames))
	}

	return Config{
		AppPort:             getEnvStr("APP_PORT", ""),
		LogLevel:            getEnvInt("LOG_LEVEL", int(slog.LevelError)),
		ShutdownPeriod:      time.Duration(getEnvInt("SHUTDOWN_PERIOD", 15)) * time.Second,
		ShutdownHardPeriod:  time.Duration(getEnvInt("SHUTDOWN_HARD_PERIOD", 3)) * time.Second,
		ReadinessDrainDelay: time.Duration(getEnvInt("READINESS_DRAIN_DELAY", 1)) * time.Second,

		RadiusSecret: getEnvStr("RADIUS_SECRET", ""),

		KeycloakClientSecret: getEnvStr("KEYCLOAK_CLIENT_SECRET", ""),

		RocketSmsLogin:         getEnvStr("ROCKET_SMS_LOGIN", ""),
		RocketSmsPass:          getEnvStr("ROCKET_SMS_PASS", ""),
		RocketSmsSender:        getEnvStr("ROCKET_SMS_SENDER", ""),
		RocketSmsTitle:         getEnvStr("ROCKET_SMS_TITLE", ""),
		RocketSmsTimeout:       time.Duration(getEnvInt("ROCKET_SMS_STATUS_TIMEOUT_SECONDS", 7)) * time.Second,
		RocketSmsTimeoutAmount: getEnvInt("ROCKET_SMS_STATUS_TIMEOUT_AMOUNT", 3),

		RedisPort:                      getEnvStr("REDIS_PORT", ""),
		RedisPass:                      getEnvStr("REDIS_PASS", ""),
		RedisFirstFactorCodeTTLMinutes: time.Duration(getEnvInt("REDIS_FIRST_FACTOR_CODE_TTL_MINUTES", 5)) * time.Minute,
		RedisFirstFactorFailedAttempts: getEnvInt("REDIS_FIRST_FACTOR_FAILED_ATTEMPTS", 3),
		RedisSecondFactorCodeTTLDays:   time.Duration(getEnvInt("REDIS_SECOND_FACTOR_CODE_TTL_DAYS", 15)) * time.Hour * 24,

		FireBasePrivateKeyPath:  getEnvStr("FIRE_BASE_PRIVATE_KEY_PATH", ""),
		PushSessionTTLInMinutes: time.Duration(getEnvInt("PUSH_SESSION_TTL_IN_MINUTES", 3)) * time.Minute,
	}
}

func getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		val, err := strconv.Atoi(value)
		if err != nil {
			return defaultVal
		}

		return val
	}

	return defaultVal
}
