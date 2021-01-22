
录制rtsp到google drive

## required
ffmpeg


## docker run
```shell
docker run \
-v {保存配置文件夹}:/camerRecord/.config \
-v {保存临时视频文件}:/camerRecord/{VideoOputPath} \
-v {保存日志文件}:/camerRecord/{.log} \
dalaomai/camer-record:latest
```