package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"path"
	"strconv"
)

var (
	goPoolSize int
)

func Load(workDir string) error {
	err := godotenv.Load(path.Join(workDir, "/.env"))
	if err != nil {
		log.Fatalf("Error loading .env file [%s]", workDir)
	}

	goPoolSize, err = strconv.Atoi(os.Getenv("GOPOOL_SIZE"))
	if err != nil {
		return err
	}

	return nil
}

func GoPoolSize() int {
	return goPoolSize
}
