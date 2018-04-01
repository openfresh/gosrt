//
//  udt_wrapper.cpp
//  VideoCast
//
//  Created by Tomohiro Matsuzawa on 2018/03/07.
//  Copyright © 2018年 CyberAgent, Inc. All rights reserved.
//

#include "udt_wrapper.h"
#include <srt/udt.h>

extern "C" {
    int udtSetLogStream(const char* logfile) {
        std::ofstream logfile_stream;
        logfile_stream.open(logfile);
        if ( !logfile_stream )
        {
            return SRT_ERROR;
        }
        else
        {
            UDT::setlogstream(logfile_stream);
        }
        
        return 0;
    }

    void logHandler_cgo(void* opaque, int level, const char* file, int line, const char* area, const char* message) {
        void logHandler(void*, int, const char*, int, const char*, const char*);
	    logHandler(opaque, level, (char*)file, line, (char*)area, (char*)message);
    }
}