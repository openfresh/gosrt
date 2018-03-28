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
}
