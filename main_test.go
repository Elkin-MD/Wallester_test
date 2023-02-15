package main

import (
	"log"
	"os"
	"testing"
	"wallester_test/models"
	"wallester_test/storage"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}

func TestSetupRoutes(t *testing.T) {
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	assert.NoError(t, err)

	err = models.MigrateCustomers(db)
	assert.NoError(t, err)
}
