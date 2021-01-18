package rtsp

import ffmpeg "github.com/u2takey/ffmpeg-go"

/*
ffmpeg -i rtsp://admin:maizhiling456@192.168.1.222:554/stream1 -c copy -f segment -strftime 1 -segment_time 10 -segment_format mkv out%Y-%m-%d_%H-%M-%S.mkv
*/

const (
	camerURL = "rtsp://admin:maizhiling456@192.168.1.222:554/stream1"
)

// Record ...
func Record(outputPath string) error {
	err := ffmpeg.Input(camerURL).Output(outputPath+"%Y-%m-%d_%H-%M-%S.mkv", ffmpeg.KwArgs{"c": "copy", "f": "segment", "strftime": 1, "segment_time": 60, "segment_format": "mkv"}).Run()
	return err
}
