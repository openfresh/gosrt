// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srt

import (
	"context"

	"github.com/openfresh/gosrt/srtapi"
)

// listenCallbackContextKey is the type of contextKeys used for listenCallback.
type listenCallbackContextKey struct{}

// WithListenCallback returns a new context.Context with the listenCallback.
func WithListenCallback(ctx context.Context, callback srtapi.SrtListenCallbackFunc) context.Context {
	return context.WithValue(ctx, listenCallbackContextKey{}, callback)
}

func listenCallbackValue(ctx context.Context) srtapi.SrtListenCallbackFunc {
	callback, _ := ctx.Value(listenCallbackContextKey{}).(srtapi.SrtListenCallbackFunc)
	return callback
}
