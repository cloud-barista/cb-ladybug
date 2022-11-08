#!/bin/sh

curl -sX POST http://localhost:1024/spider/anycall -H 'Content-Type: application/json' -d \
        '{
                "ConnectionName" : "config-aws-ap-northeast-2",
                "ReqInfo" : {
                        "FID" : "getRegionInfo"
                }
        }' | json_pp
