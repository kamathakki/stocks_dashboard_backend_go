package env

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var EnvKeys = struct {
	PG_USER            string
	PG_PASS            string
	DB_NAME            string
	BACKEND_PORT       string
	SECRET_KEY         string
	REFRESH_SECRET_KEY string
	IAM_PORT           string
	SKU_PORT           string
	WAREHOUSE_PORT     string
	REDIS_USER         string
	REDIS_PASSWORD     string
	REDIS_HOST         string
	REDIS_PORT         string
	PROTOCOL           string
	JOB_TIME_HOUR      string
	JOB_TIME_MINUTE    string
	JOB_TIME_SECOND    string
	JOB_TIME_DAY       string
	JOB_TIME_HOUR_FREQUENCY string
	JOB_TIME_MINUTE_FREQUENCY string
	JOB_TIME_SECOND_FREQUENCY string
	JOB_TIME_DAY_FREQUENCY string
	FRONTEND_PROTOCOL string
	FRONTEND_CLIENT string
}{PG_USER: "PG_USER", PG_PASS: "PG_PASS", DB_NAME: "DB_NAME", BACKEND_PORT: "BACKEND_PORT", SECRET_KEY: "SECRET_KEY",
	REFRESH_SECRET_KEY: "REFRESH_SECRET_KEY", IAM_PORT: "IAM_PORT", SKU_PORT: "SKU_PORT", WAREHOUSE_PORT: "WAREHOUSE_PORT",
	REDIS_USER: "REDIS_USER", REDIS_PASSWORD: "REDIS_PASSWORD", REDIS_HOST: "REDIS_HOST", REDIS_PORT: "REDIS_PORT", PROTOCOL: "PROTOCOL",
    JOB_TIME_HOUR: "JOB_TIME_HOUR", JOB_TIME_MINUTE: "JOB_TIME_MINUTE", JOB_TIME_SECOND: "JOB_TIME_SECOND", 
	JOB_TIME_DAY: "JOB_TIME_DAY", JOB_TIME_HOUR_FREQUENCY: "JOB_TIME_HOUR_FREQUENCY", JOB_TIME_MINUTE_FREQUENCY: "JOB_TIME_MINUTE_FREQUENCY", 
	JOB_TIME_SECOND_FREQUENCY: "JOB_TIME_SECOND_FREQUENCY", JOB_TIME_DAY_FREQUENCY: "JOB_TIME_DAY_FREQUENCY", FRONTEND_PROTOCOL: "FRONTEND_PROTOCOL", FRONTEND_CLIENT: "FRONTEND_CLIENT"}

func init() {
	err := godotenv.Load(".env.development")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	requiredEnvs := []string{"PG_USER", "PG_PASS", "DB_NAME", "BACKEND_PORT", "SECRET_KEY", "REFRESH_SECRET_KEY",
		"IAM_PORT", "SKU_PORT", "WAREHOUSE_PORT", "REDIS_USER", "REDIS_PASSWORD", "REDIS_HOST", "REDIS_PORT", "PROTOCOL",
		"JOB_TIME_HOUR", "JOB_TIME_MINUTE", "JOB_TIME_SECOND", "JOB_TIME_DAY", "JOB_TIME_HOUR_FREQUENCY", "JOB_TIME_MINUTE_FREQUENCY",
		"JOB_TIME_SECOND_FREQUENCY", "JOB_TIME_DAY_FREQUENCY", "FRONTEND_PROTOCOL", "FRONTEND_CLIENT",
	}

	for _, envVariable := range requiredEnvs {
		envVariableValue := os.Getenv(envVariable)

		if envVariableValue == "" {
			panic(fmt.Sprintf("Env variable %v is missing", envVariable))
		}
	}

}

func GetEnv[T any](key string) T {
	val, ok := os.LookupEnv(key)
	var zero T

	if !ok {
		return zero
	}

	switch any(zero).(type) {
	case int:
		i, err := strconv.Atoi(val)
		if err != nil {
			panic(fmt.Sprintf("Env %s must be int: %v", key, err))
		}
		return any(i).(T)
	case int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Env %s must be int64: %v", key, err))
		}
		return any(i).(T)
	case float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(fmt.Sprintf("Env %s must be float64: %v", key, err))
		}
		return any(f).(T)
	case bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Sprintf("Env %s must be bool: %v", key, err))
		}
		return any(b).(T)
	case string:
		return any(val).(T)
	default:
		return zero
	}
}
