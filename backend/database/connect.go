package database

import (
	"database/sql"
	"fmt"
	"github/ukilolll/trade/pkg"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	_ = pkg.LoadEnv()
)

var (
	DB_USERNAME = os.Getenv("DB_USERNAME")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	DB_HOST     = os.Getenv("DB_HOST")
	DB_NAME     = os.Getenv("DB_NAME")
)

func Connect() *sql.DB {
	var dsn = fmt.Sprintf("user=%v password=%v host=%v dbname=%v sslmode=disable", DB_USERNAME, DB_PASSWORD, DB_HOST, DB_NAME)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Panic(err)
	}
	return conn
}
