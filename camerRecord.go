//测试

package main

import (
	camerRecordClient "camerRecord/client"
	"camerRecord/rtsp"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	folderName       = "测试文件夹"
	videosPath       = "_temp/"
	uploadFileNumber = 4
)

func main() {
	var err error
	client, err := camerRecordClient.New()
	if err != nil {
		log.Fatalf("获取谷歌客户端失败:%v", err)
	}

	folderIDS, err := client.
		SearchFolder(folderName, camerRecordClient.RootFolderID)
	if err != nil {
		log.Fatalf("搜索文件夹失败:%v", err)
	}

	var parentID string
	if len(folderIDS) > 0 {
		parentID = folderIDS[0]
	} else {
		parentID, err = client.CreateFolder(folderName, []string{camerRecordClient.RootFolderID})
		if err != nil {
			log.Fatalf("创建文件夹失败:%v", err)
		}
	}

	fmt.Printf("文件夹ID:%s\n", parentID)

	go uploadFiles(client, parentID)
	rtsp.Record(videosPath)
}

func uploadFiles(client *camerRecordClient.Client, parentID string) {
	wg := sync.WaitGroup{}

	task := func(i int, filePath string) {
		defer wg.Done()

		var err error = nil
		fmt.Printf("%v 开始上传:%s\n", i, filePath)

		fileID, err := client.CreateFile(filePath, []string{parentID})
		if err == nil {
			fmt.Printf("%v 上传成功：%s  %s\n", i, filePath, fileID)
			os.Remove(filePath)
		} else {
			log.Printf("error: %v", err)
		}
	}

	for {
		files, err := ioutil.ReadDir(videosPath)
		if err != nil {
			panic(err)
		}

		taskNumber := len(files)
		if taskNumber < 2 {
			time.Sleep(30 * time.Second)
			continue
		}
		if taskNumber > uploadFileNumber {
			taskNumber = uploadFileNumber
		}
		taskNumber--

		for i := 0; i < taskNumber; i++ {
			wg.Add(1)
			go task(i, videosPath+files[i].Name())
		}

		wg.Wait()
	}
}
