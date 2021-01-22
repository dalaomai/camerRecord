
录制rtsp到google drive

## required
ffmpeg


## docker run
```shell
docker run dalaomai/camer-record:latest \
-v {保存配置文件夹}:/camerRecord/.config \
-v {保存临时视频文件}:/camerRecord/{VideoOputPath}
```