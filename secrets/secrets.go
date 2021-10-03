package secrets

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type (
	secretsMap map[string]string
)

const dataFilePathVarname string = "DATA_FILE_PATH"

var secretsFile string
var fileLock sync.Mutex

func InitFile() error {
	filePath, err := fetchSecretsFilepath()
	if err != nil {
		return err
	}

	err = createFile(filePath)
	if err != nil {
		return err
	}

	secretsFile = filePath
	return nil
}

func fetchSecretsFilepath() (string, error) {
	if filePath := os.Getenv(dataFilePathVarname); filePath != "" {
		return filePath, nil
	}
	return "", fmt.Errorf("environment variable %s has not been set; please set it to the secrets file path", dataFilePathVarname)
}

func createFile(filePath string) error {
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist, create a new empty json file
		err = os.WriteFile(filePath, []byte("{}"), 0666)
	}

	return err
}

func readJSONFile() (secretsMap, error) {
	// Read JSON file as map[string]string
	content, err := os.ReadFile(secretsFile)
	theSecrets := new(secretsMap)
	json.Unmarshal(content, theSecrets)

	return *theSecrets, err
}

func saveJSONFile(m secretsMap) error {
	content, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(secretsFile, content, 0666)
}

func writeSecret(hash string, value string) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	theSecrets, err := readJSONFile()
	if err != nil {
		return err
	}

	theSecrets[hash] = value
	err = saveJSONFile(theSecrets)

	return err
}

func LoadSecret(id string) (string, error) {
	fileLock.Lock()
	defer fileLock.Unlock()

	theSecrets, err := readJSONFile()
	if err != nil {
		return "", err
	}

	value := theSecrets[id]
	if value == "" {
		return "", nil
	}

	delete(theSecrets, id)
	err = saveJSONFile(theSecrets)

	return value, err
}

func SaveSecret(plainText string) (string, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	err := writeSecret(hash, plainText)
	return hash, err
}
