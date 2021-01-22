FROM dalaomai/camer-record-base:latest

ENV RUN_PATH /camerRecord

WORKDIR ${RUN_PATH}}
COPY camerRecord .

VOLUME [ "${RUN_PATH}/.config" , "${RUN_PATH}/_temp" ]

ENTRYPOINT ["camerRecord"]
