#!/bin/sh

curl -sX POST http://localhost:1024/spider/anycall -H 'Content-Type: application/json' -d \
        '{
                "ConnectionName" : "config-aws-ap-northeast-2",
                "ReqInfo" : {
                        "FID" : "createTags",
                        "IKeyValueList" :
			[
				{"Key":"ResourceId", "Value":"i-0ef36a4abd5b3bf8b"},
				{"Key":"Tag", "Value":"{\"Key\": \"kubernetes.io/cluster/openstack-01\", \"Value\": \"owned\"}"}
			]
                }
        }' | json_pp
