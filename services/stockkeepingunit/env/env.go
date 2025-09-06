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
	SCHEMA_NAME        string
	BACKEND_PORT       string
	SECRET_KEY         string
	REFRESH_SECRET_KEY string
	IAM_PORT           string
	SKU_PORT           string
	WAREHOUSE_PORT     string
}{PG_USER: "PG_USER", PG_PASS: "PG_PASS", DB_NAME: "DB_NAME", SCHEMA_NAME: "SCHEMA_NAME", BACKEND_PORT: "BACKEND_PORT", SECRET_KEY: "SECRET_KEY",
	REFRESH_SECRET_KEY: "REFRESH_SECRET_KEY", SKU_PORT: "SKU_PORT"}

func init() {
	err := godotenv.Load("./env/stockkeepingunit/.env.development")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	requiredEnvs := []string{"PG_USER", "PG_PASS", "DB_NAME", "SCHEMA_NAME", "BACKEND_PORT", "SECRET_KEY", "REFRESH_SECRET_KEY", "SKU_PORT"}

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
