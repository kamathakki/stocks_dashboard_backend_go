package database

import (
	"database/sql"
	"fmt"
	"stock_automation_backend_go/shared/env"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var err error

func init() {
	Open()
}

func Open() {
	connStr := fmt.Sprintf("user=%v dbname=%v password=%v sslmode=disable",
		env.GetEnv[string](env.EnvKeys.PG_USER), env.GetEnv[string](env.EnvKeys.DB_NAME), env.GetEnv[string](env.EnvKeys.PG_PASS))
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Print(err)
		panic("Panicking")
	}
	//fmt.Print("Ping response ", DB.Ping())
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func GetDB() *sql.DB {
	return DB
}
