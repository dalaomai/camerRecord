package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"camerRecord/logging"

	"github.com/dalaomai/go-onedrive/onedrive"
	"golang.org/x/oauth2"
)

var logger = logging.GetLogger()

const (
	MICROSOFT_LOGIN_BASE_URL = "https://login.microsoftonline.com"
	loginTokenURL            = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	clientID                 = "bca37e15-73ac-436c-9115-a9c6e0de67c5"
	clientSecret             = "4IM7IFP-JxmovD68t49~No1u6cHCH0I~8~"
	redirectUri              = "http://localhost:1234"
	authHttpHost             = "localhost"
	authHttpPort             = "1234"
	tokenScope               = "offline_access files.readwrite"
)

type OneDriveClient struct {
	token         string
	RefreshToken  string
	TokenData     map[string]interface{}
	TokenFilePath string
	ExpiresIn     int
	UpdateAt      int
}

func NewOneDriveClient(tokenFilePath string) (*OneDriveClient, error) {
	client := OneDriveClient{
		TokenFilePath: tokenFilePath,
	}

	err := client.readTokenData()
	if err != nil {
		client.authorize()
	}

	return &client, nil
}

func (c *OneDriveClient) tokenRefresh() error {
	/*
		POST https://login.microsoftonline.com/common/oauth2/v2.0/token
		Content-Type: application/x-www-form-urlencoded

		client_id={client_id}&redirect_uri={redirect_uri}&client_secret={client_secret}
		&refresh_token={refresh_token}&grant_type=refresh_token
	*/
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectUri)
	params.Set("client_secret", clientSecret)
	params.Set("refresh_token", c.RefreshToken)
	params.Set("grant_type", "refresh_token")

	resp, err := http.Post(
		loginTokenURL, "application/x-www-form-urlencoded",
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)

	var token_map map[string]interface{}
	if err := json.Unmarshal(body, &token_map); err != nil {
		logger.Errorf("refresh token return %s", string(body))
		return err
	}

	if _, ok := token_map["access_token"]; !ok {
		return errors.New(string(body))
	}

	c.TokenData = token_map
	err = c.extractTokenData()
	if err != nil {
		return err
	}
	c.UpdateAt = int(time.Now().Unix())

	err = c.saveTokenData()
	return err

}

func (c *OneDriveClient) GetOnedriveClient() (*onedrive.Client, error) {
	ctx := context.Background()
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return onedrive.NewClient(tc), nil
}

func (c *OneDriveClient) saveTokenData() error {
	data, _ := json.Marshal(c.TokenData)
	return ioutil.WriteFile(c.TokenFilePath, data, 0777)
}

func (c *OneDriveClient) readTokenData() error {

	token_bytes, err := ioutil.ReadFile(c.TokenFilePath)
	if err != nil {
		return err
	}
	var token_map map[string]interface{}
	err = json.Unmarshal(token_bytes, &token_map)
	if err != nil {
		return err
	}

	c.TokenData = token_map
	err = c.extractTokenData()
	return err
}

func (c *OneDriveClient) extractTokenData() error {
	if _, ok := c.TokenData["access_token"]; !ok {
		return errors.New("not access_token")
	}
	c.token = c.TokenData["access_token"].(string)

	if _, ok := c.TokenData["refresh_token"]; !ok {
		return errors.New("not refresh_token")
	}
	c.RefreshToken = c.TokenData["refresh_token"].(string)

	if _, ok := c.TokenData["expires_in"]; !ok {
		return errors.New("not expires_in")
	}

	c.ExpiresIn = int(c.TokenData["expires_in"].(float64))

	return nil
}

func (c *OneDriveClient) authorize() {
	var err error
	dataChain := make(chan map[string]interface{}, 1)
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", authHttpHost, authHttpPort),
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		if err_msg, ok := vars["error_description"]; ok {
			logger.Info(vars)
			logger.Fatal(err_msg)
		}
		if _, ok := vars["code"]; !ok {
			return
		}
		code := vars["code"][0]

		params := url.Values{}
		params.Set("client_id", clientID)
		params.Set("redirect_uri", redirectUri)
		params.Set("client_secret", clientSecret)
		params.Set("code", code)
		params.Set("grant_type", "authorization_code")
		resp, err := http.Post(
			loginTokenURL,
			"application/x-www-form-urlencoded",
			strings.NewReader(params.Encode()),
		)
		if err != nil {
			logger.Fatal(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)

		var token_map map[string]interface{}
		if err := json.Unmarshal(body, &token_map); err != nil {
			logger.Info(string(body))
			logger.Fatal(err)
		}

		dataChain <- token_map
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("scope", tokenScope)
	params.Set("redirect_uri", redirectUri)
	params.Set("response_type", "code")
	fmt.Printf("%s/common/oauth2/v2.0/authorize?%s\n", MICROSOFT_LOGIN_BASE_URL, params.Encode())

	c.TokenData = <-dataChain
	logger.Info(c.TokenData)

	err = server.Shutdown(context.TODO())
	if err != nil {
		logger.Fatal(err)
	}

	err = c.extractTokenData()
	if err != nil {
		logger.Fatal(err)
	}

	err = c.saveTokenData()
	if err != nil {
		logger.Fatal(err)
	}
}

func (c *OneDriveClient) GetToken() (string, error) {
	if int(time.Now().Unix())-60 > c.ExpiresIn+c.UpdateAt {
		err := c.tokenRefresh()
		if err != nil {
			return "", err
		}
	}

	return c.token, nil
}
