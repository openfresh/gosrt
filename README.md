[![Build Status](https://secure.travis-ci.org/openfresh/gosrt.png?branch=master)](http://travis-ci.org/openfresh/gosrt)
[![GoDoc](https://godoc.org/github.com/openfresh/gosrt/srt?status.svg)](https://godoc.org/github.com/openfresh/gosrt/srt)
[![license](https://img.shields.io/badge/license-MIT-4183c4.svg)](https://github.com/openfresh/gosrt/blob/master/LICENSE)

# gosrt

This is a [SRT](https://github.com/Haivision/srt) library for Go. This is based on the [SRT C library](https://github.com/Haivision/srt/blob/master/docs/API.md).

This library is internally binding SRT C API, but it exposes Go net package like API so that Go programmers can easy to integrate SRT into their application.

## Examples
This is a simple example that receive SRT packets from port 5000 and forwards them to localhost port 5001.

```go
l, _ := srt.Listen("srt", ":5000")
defer l.Close()
for {
    conn, _ := l.Accept()
    go func(sc net.Conn) {
        defer sc.Close()
        tc, _ := srt.Dial("srt", "127.0.0.1:5001")
        for {
            b := make([]byte, 1316)
            n, _ := sc.Read(b)
            tc.Write(b[:n])
        }
    }(conn)
}
```

## SRT Socket Options
There are several SRT socket [options](https://github.com/Haivision/srt/blob/master/docs/API.md#options). You can configure those options with a context.

For example, following code creates a context with options payloadsize to 1316 and latency to 400, then listen and dial with the options.

```go
ctx := srt.WithOptions(context.Background(), srt.Options("payloadsize", "1316", "latency", "400"))

l, err := srt.ListenContext(ctx, "srt", ":5000")

var d srt.Dialer
tc, err := d.DialContext(ctx, "srt", "127.0.0.1:5001")
```

Following table show how gosrt option corresponds to SRT C API options.

| gosrt option       | SRT C API option        |
|--------------------|-------------------------|
| transtype          | SRTO_TRANSTYPE          |
| maxbw              | SRTO_MAXBW              |
| pbkeylen           | SRTO_PBKEYLEN           |
| passphrase         | SRTO_PASSPHRASE         |
| mss                | SRTO_MSS                |
| fc                 | SRTO_FC                 |
| sndbuf             | SRTO_SNDBUF             |
| rcvbuf             | SRTO_RCVBUF             |
| ipttl              | SRTO_IPTTL              |
| iptos              | SRTO_IPTOS              |
| inputbw            | SRTO_INPUTBW            |
| oheadbw            | SRTO_OHEADBW            |
| latency            | SRTO_LATENCY            |
| tsbpdmode          | SRTO_TSBPDMODE          |
| tlpktdrop          | SRTO_TLPKTDROP          |
| snddropdelay       | SRTO_SNDDROPDELAY       |
| nakreport          | SRTO_NAKREPORT          |
| conntimeo          | SRTO_CONNTIMEO          |
| lossmaxttl         | SRTO_LOSSMAXTTL         |
| rcvlatency         | SRTO_RCVLATENCY         |
| peerlatency        | SRTO_PEERLATENCY        |
| minversion         | SRTO_MINVERSION         |
| streamid           | SRTO_STREAMID           |
| congestion         | SRTO_CONGESTION         |
| messageapi         | SRTO_MESSAGEAPI         |
| payloadsize        | SRTO_PAYLOADSIZE        |
| kmrefreshrate      | SRTO_KMREFRESHRATE      |
| kmpreannounce      | SRTO_KMPREANNOUNCE      |
| enforcedencryption | SRTO_ENFORCEDENCRYPTION |
| peeridletimeo      | SRTO_PEERIDLETIMEO      |
| packetfilter       | SRTO_PACKETFILTER       |

## Run the Example app with Docker
The example app receives SRT packets and sends them to the target address specified in .env file. In the following steps, you can send a test stream from ffmpeg to the gosrt example app, and ffplay play it. 

1. Install ffmpeg with srt support
```sh
$ brew tap homebrew-ffmpeg/ffmpeg
$ brew install homebrew-ffmpeg/ffmpeg/ffmpeg --with-srt
```

2. Run ffplay
```sh
$ ffplay -probesize 32000 -sync ext 'srt://0.0.0.0:5001?mode=listener'
```

3. Run gosrt Example app
```sh
$ cp .env.sample .env
$ docker-compose up
```

4. Run ffmpeg
```sh
$ ffmpeg -re -f lavfi -i testsrc=size=1280x720:rate=30 -f lavfi -i sine \
-vf drawtext="text='%{localtime\:%X}':fontsize=20:fontcolor=white:x=7:y=7" \
-vcodec libx264 -vb 2000k -preset ultrafast -x264-params keyint=60 \
-acodec aac -f mpegts 'srt://127.0.0.1:5000?streamid=#!::u=johnny,t=file,m=publish,r=results.csv'
```
