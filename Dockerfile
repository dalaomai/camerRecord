FROM alpine:3.7

ENV RUN_PATH /camerRecord
ENV TEMP_PATH /temp

# apk --no-cache add ca-certificates curl bash xz-libs git

WORKDIR ${TEMP_PATH}
RUN apk --no-cache add bash \
    && wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
    && tar -xf ffmpeg-release-amd64-static.tar.xz  \
    && cd ff* && mv ff* /usr/local/bin \
    && cd / && rm -rf ${TEMP_PATH}

WORKDIR ${RUN_PATH}}
# COPY dist/camerRecord_linux_amd64/camerRecord .
COPY mybin .

VOLUME [ "${RUN_PATH}/.config" , "${RUN_PATH}/_temp" ]

ENTRYPOINT ["camerRecord"]
