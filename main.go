//测试

package main

import (
	camerRecordClient "camerRecord/client"
	"camerRecord/config"
	"camerRecord/rtsp"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"camerRecord/logging"
)

var logger = logging.GetLogger()

const (
	//ConfigFolder 配置文件夹
	ConfigFolder = ".config"
)

func main() {
	config.InitConfig(config.ConfigFile)

	for _, camer := range config.Keys.Camers {
		filePath := camer.GetVideoOputPath()
		os.MkdirAll(filePath, os.ModePerm)

		go func(camer_ config.Camer) {
			for {
				err := rtsp.RecordV2(camer_.URL, filePath, camer_.VideoSegmentTime)
				logger.Error(err)
			}
		}(camer)
	}

	var err error
	var client camerRecordClient.DriveClient

	if config.Keys.Drive == "onedrive" {
		client, err = camerRecordClient.NewOnedriveServiceClient(ConfigFolder)
	} else if config.Keys.Drive == "google" {
		client, err = camerRecordClient.NewGoogleDriveClient(ConfigFolder)
	}

	if err != nil || client == nil {
		log.Fatalf("获取云盘客户端失败:%v", err)
	}

	rootFolderID, err := client.GetOrCreateFolder(config.Keys.RootFolder, camerRecordClient.RootFolderID)
	if err != nil {
		log.Fatalf("获取根“%v”文件夹失败", config.Keys.RootFolder)
	}

	uploadFiles(client, rootFolderID)

	for {
		time.Sleep(1000 * time.Second)
	}

}

type uploadFileTask struct {
	srcFile   string
	dstFileID string
}

func uploadFiles(client camerRecordClient.DriveClient, rootFolderID string) {
	uploadFiletaskLock := sync.Mutex{}
	uploadFiletaskChan := make(chan uploadFileTask, 0)
	wgLock := sync.Mutex{}
	wg := sync.WaitGroup{}

	uploadFileTaskFun := func(i int) {
		for {
			uploadFiletaskLock.Lock()
			task := <-uploadFiletaskChan
			uploadFiletaskLock.Unlock()

			var err error = nil
			logger.Debugf("%v 开始上传: %s\n", i, task.srcFile)

			fileID, err := client.CreateFile(task.srcFile, task.dstFileID)
			if err == nil {
				logger.Debugf("%v 上传成功：%s  %s\n", i, task.srcFile, fileID)
				os.Remove(task.srcFile)
			} else {
				logger.Infof("error: %v", err)
			}

			wgLock.Lock()
			wg.Done()
			wgLock.Unlock()
		}
	}

	createTaskFun := func(camer config.Camer) {
		folderID, err := client.GetOrCreateFolder(camer.Name, rootFolderID)
		if err != nil {
			logger.Fatalf("创建谷歌文件夹（%v）失败： %v", camer.Name, err)
		}

		for {
			videoOputPath := camer.GetVideoOputPath()

			files, err := ioutil.ReadDir(videoOputPath)
			if err != nil {
				panic(err)
			}

			if len(files) < 2 {
				time.Sleep(1 * time.Second)
				continue
			}

			for i := 0; i < len(files)-1; i++ {
				uploadFiletaskChan <- uploadFileTask{
					srcFile:   videoOputPath + files[i].Name(),
					dstFileID: folderID,
				}

				wgLock.Lock()
				wg.Add(1)
				wgLock.Unlock()
			}

			wg.Wait()
		}
	}

	for i := 0; i < config.Keys.ThreadNumber; i++ {
		go uploadFileTaskFun(i)
	}

	for _, camer := range config.Keys.Camers {
		go createTaskFun(camer)
	}

}
