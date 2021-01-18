package tokenutil

import (
	"context"
	"encoding/json"
	"fmt"
	"goGoogleDrive/errors"
	"log"
	"os"

	"golang.org/x/oauth2"
)

// GetTokenFromFile 从文件读取token
func GetTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("打开token文件失败：%v", err))
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return tok, errors.NewError(fmt.Sprintf("token解码失败:%v", err))
	}
	return tok, nil
}

// GetTokenFromWeb 授权获取token
func GetTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.NewError(fmt.Sprintf("Unable to read authorization code %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("Unable to retrieve token from web %v", err))
	}

	return tok, nil
}

// SaveTokenToFile 保存token到文件
func SaveTokenToFile(path string, token *oauth2.Token) {
	fmt.Printf("Saving token file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
