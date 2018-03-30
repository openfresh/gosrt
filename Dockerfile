#build stage
ARG GO_VERSION=1.10.0
FROM golang:${GO_VERSION}-alpine AS build-stage

ENV SRT_VERSION dev

RUN wget -O srt.tar.gz "https://github.com/Haivision/srt/archive/${SRT_VERSION}.tar.gz" \
    && mkdir -p /usr/src/srt \
    && tar -xzf srt.tar.gz -C /usr/src/srt --strip-components=1 \
    && rm srt.tar.gz \
    && cd /usr/src/srt \
    && apk add --no-cache --virtual .build-deps \
        ca-certificates \
        g++ \
        gcc \
        libc-dev \
        linux-headers \
  
        make \
        tcl \
        cmake \
        openssl-dev \
        tar \
    && ./configure \
    && make \
    && make install

WORKDIR /go/src/github.com/openfresh/gosrt
COPY ./ /go/src/github.com/openfresh/gosrt
RUN CGO_ENABLED=1 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o bin/livetransmit github.com/openfresh/gosrt/example \
    && go test $(go list ./... | grep -v /vendor/)

#production stage
FROM alpine:3.7

ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

CMD ["/livetransmit/bin/livetransmit"]

WORKDIR /livetransmit

RUN apk add --no-cache libstdc++ openssl

COPY --from=build-stage /go/src/github.com/openfresh/gosrt/bin/livetransmit /livetransmit/bin/
COPY --from=build-stage /usr/local/include/srt /usr/local/include/srt
COPY --from=build-stage /usr/local/lib/libsrt* /usr/local/lib/

EXPOSE 8080