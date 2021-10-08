package secrets

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type Storage interface {
	FetchSecret(id string) (string, error)
	SaveSecret(plainText string) (string, error)
}

type (
	secretsMap map[string]string

	fileStorage struct {
		secretsFile string
		fileLock    sync.Mutex
	}
)

var errEmptySecret error = fmt.Errorf("cannot save an empty secret")

func createFile(filePath string) error {
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist, create a new empty json file
		err = os.WriteFile(filePath, []byte("{}"), 0666)
	}

	return err
}

func readJSONFile(secretsFile string) (secretsMap, error) {
	// Read JSON file as map[string]string
	content, err := os.ReadFile(secretsFile)
	theSecrets := new(secretsMap)
	json.Unmarshal(content, theSecrets)

	return *theSecrets, err
}

func saveJSONFile(secretsFile string, m secretsMap) error {
	content, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(secretsFile, content, 0666)
}

func writeSecret(fs *fileStorage, hash string, value string) error {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	theSecrets, err := readJSONFile(fs.secretsFile)
	if err != nil {
		return err
	}

	theSecrets[hash] = value
	err = saveJSONFile(fs.secretsFile, theSecrets)

	return err
}

func (fs *fileStorage) FetchSecret(id string) (string, error) {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	theSecrets, err := readJSONFile(fs.secretsFile)
	if err != nil {
		return "", err
	}

	value := theSecrets[id]
	if value == "" {
		return "", nil
	}

	delete(theSecrets, id)
	err = saveJSONFile(fs.secretsFile, theSecrets)

	return value, err
}

func (fs *fileStorage) SaveSecret(plainText string) (string, error) {
	if plainText == "" {
		return "", errEmptySecret
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	err := writeSecret(fs, hash, plainText)
	return hash, err
}

func NewFileStorage(filePath string) (Storage, error) {
	err := createFile(filePath)
	if err != nil {
		return nil, err
	}

	return &fileStorage{
		secretsFile: filePath,
		fileLock:    sync.Mutex{},
	}, nil
}
