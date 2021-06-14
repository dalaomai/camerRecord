package client

type DriveClient interface {
	CreateFolder(folderName string, parent string) (id string, err error)
	CreateFile(filePath string, parent string) (id string, err error)
	GetOrCreateFolder(folderName string, parent string) (id string, err error)
}
