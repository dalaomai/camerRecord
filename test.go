package main

import (
	driveClient "camerRecord/client"
	"fmt"
	"log"
)

func main() {
	client, err := driveClient.NewOnedriveServiceClient(".config")
	if err != nil {
		log.Fatal(err)
	}

	// id, err := client.CreateFile("_temp/camer2/2021-06-11_14-23-34.mkv", "01656TG7JQJL7BPBOCAND3WWYAZ5JIKXRB")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// client.CreateFolder("camerRecord", "")
	ids, err := client.GetOrCreateFolder("testfolder", "")
	fmt.Print(err)
	fmt.Printf("%v", ids)
}
