package rtsp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/dalaomai/vdk/av"
	"github.com/dalaomai/vdk/cgo/ffmpeg"
	mp4 "github.com/dalaomai/vdk/format/mp4m"
	"github.com/dalaomai/vdk/format/rtspv2"
)

const (
	DIAL_TIMEOUT       = 3 * time.Second
	READ_WRITE_TIMEOUT = 3 * time.Second
	RECORD_TIMEOUT     = 10 * time.Second
	TIME_FORMAT        = "2006-01-02_15-04-05"
)

type AudioTranscoder struct {
	DecodeAVType av.CodecType
	EncodeAVType av.CodecType
	decoder      *ffmpeg.AudioDecoder
	encoder      *ffmpeg.AudioEncoder
}

func CreateAudioTranscoder(codecData av.AudioCodecData, encodeAVType av.CodecType) (*AudioTranscoder, error) {
	var err error
	transcoder := AudioTranscoder{
		DecodeAVType: codecData.Type(),
		EncodeAVType: encodeAVType,
	}

	transcoder.decoder, err = ffmpeg.NewAudioDecoder(codecData)
	if err != nil {
		return nil, err
	}

	transcoder.encoder, err = ffmpeg.NewAudioEncoderByCodecType(encodeAVType)
	if err != nil {
		return nil, err
	}
	err = transcoder.decoder.Setup()
	if err != nil {
		return nil, err
	}

	return &transcoder, nil
}

type Mp4Muxer struct {
	VdkMP4Muxer *mp4.Muxer
	ATranscoder *AudioTranscoder
	VideoIdx    int8
	AudioIdx    int8
}

func NewMuxer(w io.WriteSeeker) *Mp4Muxer {
	muxer := mp4.NewMuxer(w)
	return &Mp4Muxer{
		VdkMP4Muxer: muxer,
	}
}

func (muxer *Mp4Muxer) WriteHeader(streams []av.CodecData) error {
	var aTranscoder *AudioTranscoder
	var err error

	for i, stream := range streams {
		if stream.Type().IsAudio() {
			muxer.AudioIdx = int8(i)
		}
		if stream.Type().IsVideo() {
			muxer.VideoIdx = int8(i)
		}

		switch stream.Type() {
		case av.PCM_ALAW:
			aTranscoder, err = CreateAudioTranscoder(stream.(av.AudioCodecData), av.AAC)
			if err != nil {
				return err
			}
			encoderCodecData, err := aTranscoder.encoder.CodecData()
			if err != nil {
				return err
			}
			streams[i] = encoderCodecData
		}
	}
	muxer.ATranscoder = aTranscoder

	err = muxer.VdkMP4Muxer.WriteHeader(streams)
	return err
}

func (muxer *Mp4Muxer) WritePacket(pkt av.Packet) error {
	if muxer.ATranscoder != nil && pkt.Idx == muxer.AudioIdx {
		gotFrame, frame, err := muxer.ATranscoder.decoder.Decode(pkt.Data)
		if err != nil {
			return err
		}
		if !gotFrame {
			return fmt.Errorf("not get frame")
		}
		encodePkts, err := muxer.ATranscoder.encoder.Encode(frame)
		if err != nil {
			return err
		}
		if len(encodePkts) < 1 {
			// return fmt.Errorf("encode pkt not get data")
			pkt.Data = []byte{}
		} else {
			pkt.Data = encodePkts[0]
		}
	}
	return muxer.VdkMP4Muxer.WritePacket(pkt)
}

func (muxer *Mp4Muxer) WriteTrailer() error {
	if muxer.ATranscoder != nil {
		muxer.ATranscoder.encoder.Close()
		muxer.ATranscoder.decoder.Close()
	}

	return muxer.VdkMP4Muxer.WriteTrailer()
}

func RecordV2(camerURL string, outputPath string, segmentTime int) (err error) {
	client, err := rtspv2.Dial(
		rtspv2.RTSPClientOptions{
			URL:              camerURL,
			DisableAudio:     false,
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

func createMp4Muxer(outputPath string, streams []av.CodecData) (*Mp4Muxer, error) {
	fileOut, err := createOutFile(outputPath, time.Now())
	if err != nil {
		return nil, err
	}
	muxer := NewMuxer(fileOut)
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
