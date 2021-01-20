//测试

package main

import (
	camerRecordClient "camerRecord/client"
	"camerRecord/config"
	"camerRecord/rtsp"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	config.InitConfig(config.ConfigFile)

	for _, camer := range config.Keys.Camers {
		filePath := camer.GetVideoOputPath()
		os.MkdirAll(filePath, os.ModePerm)

		go rtsp.Record(camer.URL, filePath, camer.VideoSegmentTime)
	}

	var err error

	client, err := camerRecordClient.New(config.Keys.GoogleConfigFolder)
	if err != nil {
		log.Fatalf("获取谷歌客户端失败:%v", err)
	}

	rootFolderID, err := client.GetOrCreateFolder(config.Keys.GoogleFolder, camerRecordClient.RootFolderID)
	if err != nil {
		log.Fatalf("获取谷歌“%v”文件夹失败", config.Keys.GoogleFolder)
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

func uploadFiles(client *camerRecordClient.Client, rootFolderID string) {
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
			fmt.Printf("%v 开始上传:%s\n", i, task.srcFile)

			fileID, err := client.CreateFile(task.srcFile, []string{task.dstFileID})
			if err == nil {
				fmt.Printf("%v 上传成功：%s  %s\n", i, task.srcFile, fileID)
				os.Remove(task.srcFile)
			} else {
				log.Printf("error: %v", err)
			}

			wgLock.Lock()
			wg.Done()
			wgLock.Unlock()
		}
	}

	createTaskFun := func(camer config.Camer) {
		folderID, err := client.GetOrCreateFolder(camer.Name, rootFolderID)
		if err != nil {
			log.Fatalf("创建谷歌文件夹（%v）失败： %v", camer.Name, err)
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
