// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"syscall"
	"unsafe"
)

type _Socklen C.int

// SrtSocket represents SRT C API SRTSOCKET type
type SrtSocket C.SRTSOCKET

var rsa syscall.RawSockaddrAny
var rs4 syscall.RawSockaddrInet4
var rs6 syscall.RawSockaddrInet6

// Size of raw sock addr structures
const (
	SizeofSockaddrAny   = _Socklen(unsafe.Sizeof(rsa))
	SizeofSockaddrInet4 = _Socklen(unsafe.Sizeof(rs4))
	SizeofSockaddrInet6 = _Socklen(unsafe.Sizeof(rs6))
)

// SRT socket status
const (
	StatusInit       = C.SRTS_INIT
	StatusOpened     = C.SRTS_OPENED
	StatusListening  = C.SRTS_LISTENING
	StatusConnecting = C.SRTS_CONNECTING
	StatusConnected  = C.SRTS_CONNECTED
	StatusBroken     = C.SRTS_BROKEN
	StatusClosing    = C.SRTS_CLOSING
	StatusClosed     = C.SRTS_CLOSED
	StatusNonexist   = C.SRTS_NONEXIST
)

// SRT socket options
const (
	OptionMss           = C.SRTO_MSS
	OptionSndsyn        = C.SRTO_SNDSYN
	OptionRcvsyn        = C.SRTO_RCVSYN
	OptionIsn           = C.SRTO_ISN
	OptionFc            = C.SRTO_FC
	OptionSndbuf        = C.SRTO_SNDBUF
	OptionRcvbuf        = C.SRTO_RCVBUF
	OptionLinger        = C.SRTO_LINGER
	OptionUDPSndbuf     = C.SRTO_UDP_SNDBUF
	OptionUDPRcvbuf     = C.SRTO_UDP_RCVBUF
	OptionRendezvous    = C.SRTO_RENDEZVOUS
	OptionSndtimeo      = C.SRTO_SNDTIMEO
	OptionRcvtimeo      = C.SRTO_RCVTIMEO
	OptionReuseaddr     = C.SRTO_REUSEADDR
	OptionMaxbw         = C.SRTO_MAXBW
	OptionState         = C.SRTO_STATE
	OptionEvent         = C.SRTO_EVENT
	OptionSnddata       = C.SRTO_SNDDATA
	OptionRcvdata       = C.SRTO_RCVDATA
	OptionSender        = C.SRTO_SENDER
	OptionTsbpdmode     = C.SRTO_TSBPDMODE
	OptionLatency       = C.SRTO_LATENCY
	OptionTsbpddelay    = C.SRTO_TSBPDDELAY
	OptionInputbw       = C.SRTO_INPUTBW
	OptionOheadbw       = C.SRTO_OHEADBW
	OptionPassphrase    = C.SRTO_PASSPHRASE
	OptionPbkeylen      = C.SRTO_PBKEYLEN
	OptionKmstate       = C.SRTO_KMSTATE
	OptionIpttl         = C.SRTO_IPTTL
	OptionIptos         = C.SRTO_IPTOS
	OptionTlpktdrop     = C.SRTO_TLPKTDROP
	OptionSnddropdelay  = C.SRTO_SNDDROPDELAY
	OptionNakreport     = C.SRTO_NAKREPORT
	OptionVersion       = C.SRTO_VERSION
	OptionPeerversion   = C.SRTO_PEERVERSION
	OptionConntimeo     = C.SRTO_CONNTIMEO
	OptionSndkmstate    = C.SRTO_SNDKMSTATE
	OptionRcvkmstate    = C.SRTO_RCVKMSTATE
	OptionLossmaxttl    = C.SRTO_LOSSMAXTTL
	OptionRcvlatency    = C.SRTO_RCVLATENCY
	OptionPeerlatency   = C.SRTO_PEERLATENCY
	OptionMinversion    = C.SRTO_MINVERSION
	OptionStreamid      = C.SRTO_STREAMID
	OptionCongestion    = C.SRTO_CONGESTION
	OptionMessageapi    = C.SRTO_MESSAGEAPI
	OptionPayloadsize   = C.SRTO_PAYLOADSIZE
	OptionTranstype     = C.SRTO_TRANSTYPE
	OptionKmrefreshrate = C.SRTO_KMREFRESHRATE
	OptionKmpreannounce = C.SRTO_KMPREANNOUNCE
	OptionStrictenc     = C.SRTO_STRICTENC
	OptionIpv60only     = C.SRTO_IPV6ONLY
	OptionPeeridletimeo = C.SRTO_PEERIDLETIMEO
)

// SRT trans type
const (
	TypeLive    = C.SRTT_LIVE
	TypeFile    = C.SRTT_FILE
	TypeInvalid = C.SRTT_INVALID
)

// SRT log level
const (
	LogEmerg   = 0
	LogAlert   = 1
	LogFatal   = 2
	LogError   = 3
	LogWarning = 4
	LogNote    = 5
	LogInfo    = 6
	LogDebug   = 7
)

// SRT log FA
const (
	LogFAGeneral = 0
	LogFABstats  = 1
	LogFAControl = 2
	LogFAData    = 3
	LogFATsbpd   = 4
	LogFARexmit  = 5
)

// SRT log flags
const (
	LogFlagDisableTime       = 1
	LogFlagDisableThreadname = 2
	LogFlagDisableSeverity   = 4
	LogFlagDisableEOF        = 8
)

// SRT epoll opt
const (
	EpollIn  = C.SRT_EPOLL_IN
	EpollOut = C.SRT_EPOLL_OUT
	EpollErr = C.SRT_EPOLL_ERR
)

// SRT const
const (
	InvalidSock          = -1
	APIError             = -1
	DefaultSendfileBlock = 364000
	DefaultRecvfileBlock = 7280000
)
