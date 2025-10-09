package env

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var EnvKeys = struct {
	PG_HOST                   string
	PG_USER                   string
	PG_PASS                   string
	DB_NAME                   string
	BACKEND_PORT              string
	SECRET_KEY                string
	REFRESH_SECRET_KEY        string
	IAM_HOST                  string
	SKU_HOST                  string
	WAREHOUSE_HOST            string
	IAM_PORT                  string
	SKU_PORT                  string
	WAREHOUSE_PORT            string
	REDIS_USER                string
	REDIS_PASSWORD            string
	REDIS_HOST                string
	REDIS_PORT                string
	PROTOCOL                  string
	JOB_TIME_HOUR             string
	JOB_TIME_MINUTE           string
	JOB_TIME_SECOND           string
	JOB_TIME_DAY              string
	JOB_TIME_HOUR_FREQUENCY   string
	JOB_TIME_MINUTE_FREQUENCY string
	JOB_TIME_SECOND_FREQUENCY string
	JOB_TIME_DAY_FREQUENCY    string
	FRONTEND_PROTOCOL         string
	FRONTEND_CLIENT           string
}{PG_HOST: "PG_HOST", PG_USER: "PG_USER", PG_PASS: "PG_PASS", DB_NAME: "DB_NAME", BACKEND_PORT: "BACKEND_PORT", SECRET_KEY: "SECRET_KEY",
	REFRESH_SECRET_KEY: "REFRESH_SECRET_KEY", IAM_HOST: "IAM_HOST", SKU_HOST: "SKU_HOST", WAREHOUSE_HOST: "WAREHOUSE_HOST", IAM_PORT: "IAM_PORT", SKU_PORT: "SKU_PORT", WAREHOUSE_PORT: "WAREHOUSE_PORT",
	REDIS_USER: "REDIS_USER", REDIS_PASSWORD: "REDIS_PASSWORD", REDIS_HOST: "REDIS_HOST", REDIS_PORT: "REDIS_PORT", PROTOCOL: "PROTOCOL",
	JOB_TIME_HOUR: "JOB_TIME_HOUR", JOB_TIME_MINUTE: "JOB_TIME_MINUTE", JOB_TIME_SECOND: "JOB_TIME_SECOND",
	JOB_TIME_DAY: "JOB_TIME_DAY", JOB_TIME_HOUR_FREQUENCY: "JOB_TIME_HOUR_FREQUENCY", JOB_TIME_MINUTE_FREQUENCY: "JOB_TIME_MINUTE_FREQUENCY",
	JOB_TIME_SECOND_FREQUENCY: "JOB_TIME_SECOND_FREQUENCY", JOB_TIME_DAY_FREQUENCY: "JOB_TIME_DAY_FREQUENCY", FRONTEND_PROTOCOL: "FRONTEND_PROTOCOL", FRONTEND_CLIENT: "FRONTEND_CLIENT"}

func init() {
	ENV := os.Getenv("ENV")
	var rootDIR string = "."
	if ENV == "production" {
		rootDIR = "/app"
	}
	envPath := rootDIR + "/.env." + ENV
	fmt.Println("envPath", envPath)
        if err := godotenv.Load(envPath); err != nil {
            log.Printf("warning: failed to load env file %s: %v", envPath, err)
        }
    // Load secrets file from shared/env by default (overridable via SECRET_PATH)
    var secretPath string = ""
    if secretPath == "" {
        if ENV == "production" {
            secretPath = "/run/secrets/secret.json"
        } else {
            secretPath = "shared/env/secret.json"
        }
    }
    loadSecrets(secretPath)
	requiredEnvs := []string{"PG_HOST", "PG_USER", "PG_PASS", "DB_NAME", "BACKEND_PORT", "SECRET_KEY", "REFRESH_SECRET_KEY",
		"IAM_HOST", "SKU_HOST", "WAREHOUSE_HOST", "IAM_PORT", "SKU_PORT", "WAREHOUSE_PORT", "REDIS_USER", "REDIS_PASSWORD", "REDIS_HOST", "REDIS_PORT", "PROTOCOL",
		"JOB_TIME_HOUR", "JOB_TIME_MINUTE", "JOB_TIME_SECOND", "JOB_TIME_DAY", "JOB_TIME_HOUR_FREQUENCY", "JOB_TIME_MINUTE_FREQUENCY",
		"JOB_TIME_SECOND_FREQUENCY", "JOB_TIME_DAY_FREQUENCY", "FRONTEND_PROTOCOL", "FRONTEND_CLIENT",
	}

	for _, envVariable := range requiredEnvs {
		envVariableValue := os.Getenv(envVariable)

		if envVariableValue == "" {
			panic(fmt.Sprintf("Env variable %v is missings", envVariable))
		}
	}

}

var secretValues = map[string]string{}

func loadSecrets(path string) {
    data, err := os.ReadFile(path)
    if err != nil {
        log.Printf("warning: failed to read secrets file %s: %v", path, err)
        return
    }
    var parsed map[string]any
    if err := json.Unmarshal(data, &parsed); err != nil {
        log.Printf("warning: failed to parse secrets file %s: %v", path, err)
        return
    }
    tmp := make(map[string]string, len(parsed))
    for k, v := range parsed {
        switch vv := v.(type) {
        case string:
            tmp[k] = vv
        case float64:
            // JSON numbers are float64; convert to int-like string when possible
            if float64(int64(vv)) == vv {
                tmp[k] = fmt.Sprintf("%d", int64(vv))
            } else {
                tmp[k] = fmt.Sprintf("%v", vv)
            }
        case bool:
            tmp[k] = fmt.Sprintf("%t", vv)
        default:
            tmp[k] = fmt.Sprintf("%v", vv)
        }
    }
    secretValues = tmp
}

func GetEnv[T any](key string) T {
    // Prefer secret value if present
    if v, ok := secretValues[key]; ok {
        return coerceEnvValue[T](key, v)
    }
    val, ok := os.LookupEnv(key)
	var zero T

	if !ok {
		return zero
	}

    return coerceEnvValue[T](key, val)
}

func coerceEnvValue[T any](key string, val string) T {
    var zero T
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
