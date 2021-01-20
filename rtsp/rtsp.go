package rtsp

import (
	"fmt"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

/*
ffmpeg -i rtsp://admin:xxxxx@192.168.1.222:554/stream1 -c copy -f segment -strftime 1 -segment_time 10 -segment_format mkv out%Y-%m-%d_%H-%M-%S.mkv
*/

// Record ...
func Record(camerURL string, outputPath string, segmentTime int) error {
	err := ffmpeg.Input(camerURL).Output(outputPath+"%Y-%m-%d_%H-%M-%S.mkv", ffmpeg.KwArgs{"c": "copy", "f": "segment", "strftime": 1, "segment_time": segmentTime, "segment_format": "mkv"}).Run()
	fmt.Printf("%v %v %v", err, camerURL, outputPath)
	return err
}
