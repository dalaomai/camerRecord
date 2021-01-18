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
	go rtsp.Record(videosPath)

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

	uploadFiles(client, parentID)
}

func uploadFiles(client *camerRecordClient.Client, parentID string) {
	taskLock := sync.Mutex{}
	taskChan := make(chan string, 0)
	wgLock := sync.Mutex{}
	wg := sync.WaitGroup{}

	task := func(i int) {
		for {
			taskLock.Lock()
			filePath := <-taskChan
			taskLock.Unlock()

			var err error = nil
			fmt.Printf("%v 开始上传:%s\n", i, filePath)

			fileID, err := client.CreateFile(filePath, []string{parentID})
			if err == nil {
				fmt.Printf("%v 上传成功：%s  %s\n", i, filePath, fileID)
				os.Remove(filePath)
			} else {
				log.Printf("error: %v", err)
			}

			wgLock.Lock()
			wg.Done()
			wgLock.Unlock()
		}
	}

	for i := 0; i < uploadFileNumber; i++ {
		go task(i)
	}

	for {
		files, err := ioutil.ReadDir(videosPath)
		if err != nil {
			panic(err)
		}

		if len(files) < 2 {
			time.Sleep(1 * time.Second)
			continue
		}

		for i := 0; i < len(files)-1; i++ {
			log.Printf("debug:%s %v", files[i].Name(), i)
			taskChan <- videosPath + files[i].Name()

			wgLock.Lock()
			wg.Add(1)
			wgLock.Unlock()
		}

		wg.Wait()
	}
}
