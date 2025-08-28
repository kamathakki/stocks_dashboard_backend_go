package env

import (
	"fmt"
	"log"
	"os"

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
}{PG_USER: "PG_USER", PG_PASS: "PG_PASS", DB_NAME: "DB_NAME", SCHEMA_NAME: "SCHEMA_NAME", BACKEND_PORT: "BACKEND_PORT",
	SECRET_KEY: "SECRET_KEY", REFRESH_SECRET_KEY: "REFRESH_SECRET_KEY", IAM_PORT: "IAM_PORT"}

func init() {
	err := godotenv.Load("./env/iam/.env.development")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	requiredEnvs := []string{"PG_USER", "PG_PASS", "DB_NAME", "SCHEMA_NAME", "BACKEND_PORT", "SECRET_KEY", "REFRESH_SECRET_KEY", "IAM_PORT"}

	for _, envVariable := range requiredEnvs {
		envVariableValue := os.Getenv(envVariable)

		if envVariableValue == "" {
			panic(fmt.Sprintf("Env variable %v is missing", envVariable))
		}
	}
}

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)

	if !ok {
		return ""
	}

	return val
}
