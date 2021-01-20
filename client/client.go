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
	credFileName  = "credentials.json"
	tokenFileName = "token.json"

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

// GetOrCreateFolder 获取文件夹ID
func (c Client) GetOrCreateFolder(folderName string, parent string) (id string, err error) {
	ids, err := c.SearchFolder(folderName, parent)
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}

	id, err = c.CreateFolder(folderName, []string{parent})
	if err != nil {
		return "", nil
	}

	return id, err
}

// New 获取client
func New(configFolder string) (*Client, error) {
	cred, err := ioutil.ReadFile(configFolder + credFileName)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("谷歌API配置文件读取失败: %v", err))
	}

	// 获取全部权限
	config, err := google.ConfigFromJSON(cred, drive.DriveScope)
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("谷歌API配置文件解析失败: %v", err))
	}

	client, err := getOauthClient(config, configFolder)
	if err != nil {
		return nil, err
	}

	service, err := drive.New(client)
	return &Client{service}, err
}

func getOauthClient(config *oauth2.Config, configFolder string) (*http.Client, error) {
	var token *oauth2.Token
	var err error

	token, err = tokenutil.GetTokenFromFile(configFolder + tokenFileName)
	if err != nil {
		log.Println(err)

		token, err = tokenutil.GetTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		tokenutil.SaveTokenToFile(configFolder+tokenFileName, token)
	}
	return config.Client(context.Background(), token), err
}
