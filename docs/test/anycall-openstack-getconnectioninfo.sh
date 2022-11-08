#!/bin/sh

curl -sX POST http://localhost:1024/spider/anycall -H 'Content-Type: application/json' -d \
        '{
                "ConnectionName" : "config-openstack-regionone",
                "ReqInfo" : {
                        "FID" : "getConnectionInfo"
                }
        }' | json_pp
