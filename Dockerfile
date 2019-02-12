FROM golang:1.11-alpine AS build-env

RUN apk update && apk add git make g++ bash

ENV PUSHER_TESTER_PATH $GOPATH/src/github.com/topfreegames/pusher-tester

COPY . $PUSHER_TESTER_PATH

RUN make -C $PUSHER_TESTER_PATH build && \
    mkdir -p /etc/pusher_tester /var/pusher_tester /app /app/config && \
    mv $PUSHER_TESTER_PATH/bin/pusher-tester /app/pusher-tester && \
    mv $PUSHER_TESTER_PATH/config/default.yaml /app/config/default.yaml

CMD ["/app/pusher-tester", "--cfgFile", "/app/config/default.yaml"]