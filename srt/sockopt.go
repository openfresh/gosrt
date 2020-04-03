// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srt

import (
	"context"
	"strconv"

	"github.com/openfresh/gosrt/srtapi"
)

const (
	typeString = 0 + iota
	typeInt
	typeInt64
	typeBool
)

const (
	bindPre = 0 + iota
	bindPost
)

type socketOption struct {
	name    string
	proto   int
	sym     int
	binding int
	typ     int
}

func (o *socketOption) apply(s int, v string) error {
	ov, err := o.extract(v)
	if err != nil {
		return err
	}
	switch ov := ov.(type) {
	case string:
		return srtapi.SetsockoptString(s, 0, o.sym, ov)
	case int:
		return srtapi.SetsockoptInt(s, 0, o.sym, ov)
	case int64:
		return srtapi.SetsockoptInt64(s, 0, o.sym, ov)
	case bool:
		return srtapi.SetsockoptBool(s, 0, o.sym, ov)
	}
	return nil
}

func (o *socketOption) extract(v string) (ov interface{}, err error) {
	switch o.typ {
	case typeString:
		ov = v
	case typeInt:
		ov, err = strconv.Atoi(v)
	case typeInt64:
		ov, err = strconv.ParseInt(v, 10, 64)
	case typeBool:
		ov, err = strconv.ParseBool(v)
	}
	return
}

var srtOptions = []socketOption{
	{"transtype", 0, srtapi.OptionTranstype, bindPre, typeInt},
	{"maxbw", 0, srtapi.OptionMaxbw, bindPre, typeInt64},
	{"pbkeylen", 0, srtapi.OptionPbkeylen, bindPre, typeInt},
	{"passphrase", 0, srtapi.OptionPassphrase, bindPre, typeString},

	{"mss", 0, srtapi.OptionMss, bindPre, typeInt},
	{"fc", 0, srtapi.OptionFc, bindPre, typeInt},
	{"sndbuf", 0, srtapi.OptionSndbuf, bindPre, typeInt},
	{"rcvbuf", 0, srtapi.OptionRcvbuf, bindPre, typeInt},
	{"ipttl", 0, srtapi.OptionIpttl, bindPre, typeInt},
	{"iptos", 0, srtapi.OptionIptos, bindPre, typeInt},
	{"inputbw", 0, srtapi.OptionInputbw, bindPost, typeInt64},
	{"oheadbw", 0, srtapi.OptionOheadbw, bindPost, typeInt},
	{"latency", 0, srtapi.OptionLatency, bindPre, typeInt},
	{"tsbpdmode", 0, srtapi.OptionTsbpdmode, bindPre, typeBool},
	{"tlpktdrop", 0, srtapi.OptionTlpktdrop, bindPre, typeBool},
	{"snddropdelay", 0, srtapi.OptionSnddropdelay, bindPost, typeInt},
	{"nakreport", 0, srtapi.OptionNakreport, bindPre, typeBool},
	{"conntimeo", 0, srtapi.OptionConntimeo, bindPre, typeInt},
	{"lossmaxttl", 0, srtapi.OptionLossmaxttl, bindPre, typeInt},
	{"rcvlatency", 0, srtapi.OptionRcvlatency, bindPre, typeInt},
	{"peerlatency", 0, srtapi.OptionPeerlatency, bindPre, typeInt},
	{"minversion", 0, srtapi.OptionMinversion, bindPre, typeInt},
	{"streamid", 0, srtapi.OptionStreamid, bindPre, typeString},
	{"congestion", 0, srtapi.OptionCongestion, bindPre, typeString},
	{"messageapi", 0, srtapi.OptionMessageapi, bindPre, typeBool},
	{"payloadsize", 0, srtapi.OptionPayloadsize, bindPre, typeInt},
	{"kmrefreshrate", 0, srtapi.OptionKmrefreshrate, bindPre, typeInt},
	{"kmpreannounce", 0, srtapi.OptionKmpreannounce, bindPre, typeInt},
	{"enforcedencryption", 0, srtapi.OptionEnforcedencryption, bindPre, typeBool},
	{"peeridletimeo", 0, srtapi.OptionPeeridletimeo, bindPre, typeInt},
	{"packetfilter", 0, srtapi.OptionPacketfilter, bindPre, typeString},
}

type option struct {
	key   string
	value string
}

// OptionSet is a set of options.
type OptionSet struct {
	list []option
}

// optionContextKey is the type of contextKeys used for options.
type optionContextKey struct{}

func optionValue(ctx context.Context) optionMap {
	options, _ := ctx.Value(optionContextKey{}).(*optionMap)
	if options == nil {
		return optionMap(nil)
	}
	return *options
}

// optionMap is the representation of the option set held in the context type.
type optionMap map[string]string

// WithOptions returns a new context.Context with the given options added.
// A option overwrites a prior option with the same key.
func WithOptions(ctx context.Context, options OptionSet) context.Context {
	childOptions := make(optionMap)
	parentOptions := optionValue(ctx)

	for k, v := range parentOptions {
		childOptions[k] = v
	}
	for _, option := range options.list {
		childOptions[option.key] = option.value
	}
	return context.WithValue(ctx, optionContextKey{}, &childOptions)
}

// Options takes an even number of strings representing key-value pairs
// and makes a OptionSet containing them.
// A option overwrites a prior option with the same key.
func Options(args ...string) OptionSet {
	if len(args)%2 != 0 {
		panic("uneven number of arguments to gosrt.Options")
	}
	options := OptionSet{}
	for i := 0; i+1 < len(args); i += 2 {
		options.list = append(options.list, option{key: args[i], value: args[i+1]})
	}
	return options
}

// Option returns the value of the option with the given key on ctx, and a boolean indicating
// whether that option exists.
func Option(ctx context.Context, key string) (string, bool) {
	ctxOptions := optionValue(ctx)
	v, ok := ctxOptions[key]
	return v, ok
}

func configure(ctx context.Context, s int, binding int) error {
	ctxOptions := optionValue(ctx)
	for _, o := range srtOptions {
		if o.binding == binding {
			if v, ok := ctxOptions[o.name]; ok {
				if err := o.apply(s, v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
