package client

import (
	"context"
	"fmt"

	"github.com/dalaomai/go-onedrive/onedrive"
)

type OnedriveServiceClient struct {
	client *OneDriveClient
	drive  *onedrive.Drive
}

func NewOnedriveServiceClient(tokenFolderPath string) (*OnedriveServiceClient, error) {
	client, err := NewOneDriveClient(fmt.Sprintf("%s/onedrive.json", tokenFolderPath))
	if err != nil {
		return nil, err
	}
	serviceClient := OnedriveServiceClient{client: client}
	serviceClient.getDrive()
	return &serviceClient, nil
}

func (sc *OnedriveServiceClient) getDriveClient() (*onedrive.Client, error) {
	return sc.client.GetOnedriveClient()
}

func (sc *OnedriveServiceClient) getDrive() error {
	c, err := sc.getDriveClient()
	if err != nil {
		return err
	}

	rsp, err := c.Drives.List(context.TODO())
	if err != nil {
		return err
	}

	sc.drive = rsp.Drives[0]

	return nil
}

func (sc *OnedriveServiceClient) GetDirve() *onedrive.Drive {
	return sc.drive
}

func (sc *OnedriveServiceClient) CreateFolder(folderName string, parent string) (id string, err error) {
	c, err := sc.getDriveClient()
	if err != nil {
		return "", err
	}

	itme, err := c.DriveItems.CreateNewFolder(context.TODO(), sc.drive.Id, parent, folderName)
	if err != nil {
		return "", err
	}

	return itme.Id, nil
}

func (sc *OnedriveServiceClient) CreateFile(filePath string, parent string) (id string, err error) {
	c, err := sc.getDriveClient()
	if err != nil {
		return "", err
	}
	item, err := c.DriveItems.UploadLargeFile(context.TODO(), sc.drive.Id, parent, filePath)
	if err != nil {
		return "", err
	}
	return item.Id, nil
}

func (sc *OnedriveServiceClient) SearchFolder(folderName string, parent string, strict bool) (ids []string, err error) {
	c, err := sc.getDriveClient()
	if err != nil {
		return nil, err
	}
	if strict {
		folderName = fmt.Sprintf("{%s}", folderName)
	}

	rsp, err := c.DriveSearch.SearchWithParent(context.TODO(), parent, folderName)
	if err != nil {
		return nil, err
	}

	ids = []string{}
	for i := 0; i < len(rsp.DriveItems); i++ {
		id := rsp.DriveItems[i].Id
		if id != parent {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func (sc *OnedriveServiceClient) GetOrCreateFolder(folderName string, parent string) (id string, err error) {
	ids, err := sc.SearchFolder(folderName, parent, true)
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}

	id, err = sc.CreateFolder(folderName, parent)
	return id, err
}

func (sc *OnedriveServiceClient) CreateLargeFile(filePath string, parent string) (interface{}, error) {
	c, err := sc.getDriveClient()
	if err != nil {
		return "", err
	}
	item, err := c.DriveItems.UploadLargeFile(context.TODO(), sc.drive.Id, parent, filePath)
	if err != nil {
		return "", err
	}
	return item, nil
}
