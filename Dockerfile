FROM golang:1.11-alpine AS build-env

RUN apk update && apk add git make

ENV PUSHER_TESTER_PATH $GOPATH/src/github.com/topfreegames/pusher-tester
ENV LIBRDKAFKA_VERSION 0.11.5
ENV CPLUS_INCLUDE_PATH /usr/local/include

COPY . $PUSHER_TESTER_PATH

RUN apk add --no-cache make git g++ bash python wget && \
    wget -O /root/librdkafka-${LIBRDKAFKA_VERSION}.tar.gz https://github.com/edenhill/librdkafka/archive/v${LIBRDKAFKA_VERSION}.tar.gz && \
    tar -xzf /root/librdkafka-${LIBRDKAFKA_VERSION}.tar.gz -C /root && \
    cd /root/librdkafka-${LIBRDKAFKA_VERSION} && \
    ./configure && make && make install && make clean && ./configure --clean

RUN make -C $PUSHER_TESTER_PATH build && \
    mkdir -p /etc/pusher_tester /var/pusher_tester /app /app/config && \
    mv $PUSHER_TESTER_PATH/bin/pusher-tester /app/pusher-tester && \
    mv $PUSHER_TESTER_PATH/config/default.yaml /app/config/default.yaml 

CMD ["/app/pusher-tester", "--cfgFile", "/app/config/default.yaml"]