package main

import (
	"fmt"
	"os"

	"github.com/folcoz/milestone1-code/secrets"
	"github.com/folcoz/milestone1-code/server"
)

const dataFilePathVarname string = "DATA_FILE_PATH"

func initStorage() (secrets.Storage, error) {
	filePath, err := fetchSecretsFilepath()
	if err != nil {
		return nil, err
	}

	return secrets.NewFileStorage(filePath)
}

func fetchSecretsFilepath() (string, error) {
	if filePath := os.Getenv(dataFilePathVarname); filePath != "" {
		return filePath, nil
	}
	return "", fmt.Errorf("environment variable %s has not been set; please set it to the secrets file path", dataFilePathVarname)
}

func main() {
	storage, err := initStorage()
	if err != nil {
		quit(err, 1)
	}

	server.StartListener("0.0.0.0:8080", storage)
}

func quit(err error, exitCode int) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(exitCode)
}
