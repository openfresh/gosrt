// Copyright (c) 2020 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

/*

#include <srt/srt.h>

// The gateway function
int SrtListenCallback_cgo(void* opaq, SRTSOCKET ns, int hsversion,
    const struct sockaddr* peeraddr, const char* streamid)
{
	int srtListenCallback(void*, SRTSOCKET, int, const struct sockaddr*, const char*);
	return srtListenCallback(opaq, ns, hsversion, peeraddr, streamid);
}
*/
import "C"
