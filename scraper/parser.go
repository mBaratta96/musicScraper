package scraper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type ConfigData struct {
	Delay        int  `json:"request_delay"`
	Authenticate bool `json:"authenticate"`
	SaveCookies  bool `json:"save_cookies"`
}

type RYMCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type JsonData interface {
	ConfigData | []RYMCookie
}

func ReadUserConfiguration(path string) (ConfigData, error) {
	return readJsonFile[ConfigData](path)
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Cookie file not found. Authentication is required.")
	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func readJsonFile[D JsonData](path string) (D, error) {
	var configData D
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("No login file found in " + path + ". Skipped user authentication")
		return configData, errors.New("No file")
	}
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Error when opening file: ")
	}
	err = json.Unmarshal(configFile, &configData)
	if err != nil {
		panic(err)
	}
	return configData, nil
}

func ReadCookie(path string) map[string]string {
	cookies, err := readJsonFile[[]RYMCookie](path)
	if err != nil {
		panic(err)
	}
	cookieMap := make(map[string]string)
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}
	return cookieMap
}

func SaveCookie(cookies map[string]string, path string) {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	as_json, err := json.MarshalIndent(cookies, "", "\t")
	if err != nil {
		panic(err)
	}
	f.Write(as_json)
}
