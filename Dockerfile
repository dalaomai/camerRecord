# FROM alpine:3.14
FROM scratch

ENV RUN_PATH /camerRecord

WORKDIR ${RUN_PATH}
COPY camerRecord .
COPY .config_template/config.json .config/config.json

VOLUME [ "${RUN_PATH}/.config"  ]

ENTRYPOINT ["./camerRecord"]
