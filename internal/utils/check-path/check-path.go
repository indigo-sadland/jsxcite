package check_path

import (
	"net/url"
	"os"
	"strings"
)

func IsLocalPath(path string) (bool, error) {
	var err error

	if strings.HasPrefix(path, "~") {
		path, err = ExpandPath(path)
		if err != nil {
			return false, err
		}
	}

	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, err
	}

	return true, nil
}

func IsURL(path string) bool {

	_, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return true
}

// ExpandPath handles '~/' path
func ExpandPath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path = homeDir + path[1:]
	return path, err
}
