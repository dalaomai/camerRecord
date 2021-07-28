package rtsp

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/deepch/vdk/av"
	mp4 "github.com/deepch/vdk/format/mp4m"
	"github.com/deepch/vdk/format/rtspv2"
)

const (
	DIAL_TIMEOUT       = 3 * time.Second
	READ_WRITE_TIMEOUT = 3 * time.Second
	RECORD_TIMEOUT     = 10 * time.Second
	TIME_FORMAT        = "2006-01-02_15-04-05"
)

func RecordV2(camerURL string, outputPath string, segmentTime int) (err error) {
	client, err := rtspv2.Dial(
		rtspv2.RTSPClientOptions{
			URL:              camerURL,
			DisableAudio:     true,
			DialTimeout:      DIAL_TIMEOUT,
			ReadWriteTimeout: READ_WRITE_TIMEOUT,
			Debug:            false,
		},
	)
	if err != nil {
		return err
	}
	defer client.Close()

	mp4Muxer, err := createMp4Muxer(outputPath, client.CodecData)
	if err != nil {
		return err
	}

	segmentTimer := time.NewTicker(time.Duration(segmentTime) * time.Second)
	defer segmentTimer.Stop()

	recordTimer := time.NewTimer(RECORD_TIMEOUT)
	defer recordTimer.Stop()

re:
	for {
		select {
		case <-segmentTimer.C:
			err = mp4Muxer.WriteTrailer()
			if err != nil {
				return err
			}

			mp4Muxer, err = createMp4Muxer(outputPath, client.CodecData)
			if err != nil {
				return err
			}
		case pck := <-client.OutgoingPacketQueue:
			recordTimer.Reset(RECORD_TIMEOUT)

			err := mp4Muxer.WritePacket(*pck)
			if err != nil {
				break re
			}
		case signals := <-client.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				mp4Muxer.WriteTrailer()
				mp4Muxer, err = createMp4Muxer(outputPath, client.CodecData)
				if err != nil {
					return err
				}
			case rtspv2.SignalStreamRTPStop:
				err = fmt.Errorf("receive stop signal:%v", signals)
				break re
			}
		case <-recordTimer.C:
			err = errors.New("record timeout")
			break re
		}

	}
	mp4Muxer.WriteTrailer()
	return
}

func createMp4Muxer(outputPath string, streams []av.CodecData) (*mp4.Muxer, error) {
	fileOut, err := createOutFile(outputPath, time.Now())
	if err != nil {
		return nil, err
	}
	muxer := mp4.NewMuxer(fileOut)
	err = muxer.WriteHeader(streams)
	if err != nil {
		return nil, err
	}
	return muxer, nil
}

func createOutFile(outputPath string, now time.Time) (*os.File, error) {
	fileOut, err := os.Create(
		path.Join(
			outputPath, fmt.Sprintf("%v.mp4", now.Format(TIME_FORMAT)),
		),
	)
	if err != nil {
		return nil, err
	}
	return fileOut, nil
}

func CreateRTSPClient(url string) (*rtspv2.RTSPClient, error) {
	RTSPClient, err := rtspv2.Dial(
		rtspv2.RTSPClientOptions{
			URL:              url,
			DisableAudio:     false,
			DialTimeout:      3 * time.Second,
			ReadWriteTimeout: 2 * time.Second,
			Debug:            false,
		},
	)
	if err != nil {
		return nil, err
	}

	return RTSPClient, nil
}
