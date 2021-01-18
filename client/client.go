//谷歌云盘client

package client

import (
	"camerRecord/errors"
	"camerRecord/tokenutil"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const (
	credFile  = "config/credentials.json"
	tokenFile = "config/token.json"

	// RootFolderID ...
	RootFolderID = "root"
)

// Client 客户端
type Client struct {
	service *drive.Service
}

// PrintFiles 获取文件列表
func (c Client) PrintFiles() (string, error) {
	r, err := c.service.Files.List().PageSize(50).Corpora("user").
		Q("mimeType='application/vnd.google-apps.folder' and 'root' in parents").
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		return "", err
	}

	var result string
	for index, file := range r.Files {
		result += fmt.Sprintf("%v  %s (%s)\n", index, file.Name, file.Id)
	}
	return result, nil
}

// CreateFolder 创建文件夹
func (c Client) CreateFolder(folderName string, parents []string) (id string, err error) {
	file, err := c.service.Files.Create(&drive.File{MimeType: "application/vnd.google-apps.folder", Name: folderName, Parents: parents}).Do()
	return file.Id, err
}

// CreateFile 创建文件
func (c Client) CreateFile(filePath string, parents []string) (id string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := filepath.Base(filePath)
	//MimeType: "application/vnd.google-apps.file",
	newFile, err := c.service.Files.Create(&drive.File{Name: filename, Parents: parents}).Media(file).Do()
	if err != nil {
		return "", err
	}

	return newFile.Id, nil
}

// SearchFolder 查找文件夹
func (c Client) SearchFolder(folderName string, parent string) (ids []string, err error) {

	parent = "'" + parent + "'"
	folderName = "'" + folderName + "'"

	r, err := c.service.Files.List().PageSize(50).Corpora("user").
		Q("mimeType='application/vnd.google-apps.folder' and trashed=false and " + parent + " in parents and " + "name = " + folderName).
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		return nil, err
	}

	ids = make([]string, 0)
	for _, f := range r.Files {
		ids = append(ids, f.Id)
	}
	return ids, nil
}

// New 获取client
func New() (*Client, error) {
	cred, err := ioutil.ReadFile(credFile)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("谷歌API配置文件读取失败: %v", err))
	}

	// 获取全部权限
	config, err := google.ConfigFromJSON(cred, drive.DriveScope)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("谷歌API配置文件解析失败: %v", err))
	}

	client, err := getOauthClient(config)
	if err != nil {
		return nil, err
	}

	service, err := drive.New(client)
	return &Client{service}, err
}

func getOauthClient(config *oauth2.Config) (*http.Client, error) {
	var token *oauth2.Token
	var err error

	token, err = tokenutil.GetTokenFromFile(tokenFile)
	if err != nil {
		log.Println(err)

		token, err = tokenutil.GetTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		tokenutil.SaveTokenToFile(tokenFile, token)
	}
	return config.Client(context.Background(), token), err
}
