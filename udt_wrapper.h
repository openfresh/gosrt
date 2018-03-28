//
//  udt_wrapper.h
//  VideoCast
//
//  Created by Tomohiro Matsuzawa on 2018/03/07.
//  Copyright © 2018年 CyberAgent, Inc. All rights reserved.
//

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