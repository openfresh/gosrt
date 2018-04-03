// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

#ifndef udt_wrapper_h
#define udt_wrapper_h

#include <srt/srt.h>

#ifdef __cplusplus
extern "C" {
#endif
    
int udtSetLogStream(const char* logfile);

#ifdef __cplusplus
}
#endif

#endif /* udt_wrapper_h */