FROM golang:1.15.6-alpine


#
# Tiny Container that holds ffmpeg, ffprobe and can build / run golang programs.
#
#
RUN apk --no-cache add ca-certificates curl bash xz-libs git
WORKDIR /tmp
RUN curl -L -O https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
RUN tar -xf ffmpeg-release-amd64-static.tar.xz && \
    cd ff* && mv ff* /usr/local/bin

WORKDIR /

ENTRYPOINT ["/bin/bash"]