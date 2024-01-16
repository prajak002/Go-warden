package database

import (
	"crypto/tls"
	"fmt"
	"os"

	mysqlTiDB "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitializeDB() (*gorm.DB, error) {
	DB_URL := os.Getenv("DB_URL")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	if DB_URL == "" || DB_PORT == "" || DB_NAME == "" || DB_USER == "" || DB_PASSWORD == "" {
		return nil, fmt.Errorf("failed to read environment variables for database")
	}

	mysqlTiDB.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: DB_URL,
	})

	dsn := DB_USER + ":" + DB_PASSWORD + "@tcp(" + DB_URL + ":" + DB_PORT + ")/" + DB_NAME + "?tls=tidb&parseTime=true"
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
