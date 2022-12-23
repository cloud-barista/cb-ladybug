#!/bin/bash
# -----------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then
	echo "./cluster-create.sh <namespace> <clsuter name> <service type>"
	echo "./cluster-create.sh cb-ladybug-ns cluster-01 <multi or single>"
	exit 0;
fi

source ./conf.env

# ------------------------------------------------------------------------------
# const


# -----------------------------------------------------------------
# parameter

# 1. namespace
if [ "$#" -gt 0 ]; then v_NAMESPACE="$1"; else	v_NAMESPACE="${NAMESPACE}"; fi
if [ "${v_NAMESPACE}" == "" ]; then
	read -e -p "Namespace ? : " v_NAMESPACE
fi
if [ "${v_NAMESPACE}" == "" ]; then echo "[ERROR] missing <namespace>"; exit -1; fi

# 2. Cluster Name
if [ "$#" -gt 1 ]; then v_CLUSTER_NAME="$2"; else	v_CLUSTER_NAME="${CLUSTER_NAME}"; fi
if [ "${v_CLUSTER_NAME}" == "" ]; then
	read -e -p "Cluster name  ? : "  v_CLUSTER_NAME
fi
if [ "${v_CLUSTER_NAME}" == "" ]; then echo "[ERROR] missing <cluster name>"; exit -1; fi

# 3. Service Type
if [ "$#" -gt 2  ]; then v_SERVICE_TYPE="$3"; else	v_SERVICE_TYPE="${SERVICE_TYPE}"; fi
if [ "${v_SERVICE_TYPE}" == ""  ]; then
	read -e -p "Service Type  ? : "  v_SERVICE_TYPE
fi
if [ "${v_SERVICE_TYPE}" == ""  ]; then echo "[ERROR] missing <service type>"; exit -1; fi


c_URL_LADYBUG_NS="${c_URL_LADYBUG}/ns/${v_NAMESPACE}"


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Namespace                  is '${v_NAMESPACE}'"
echo "- Cluster name               is '${v_CLUSTER_NAME}'"
echo "- Service type               is '${v_SERVICE_TYPE}'"

# ------------------------------------------------------------------------------
# Create a cluster
create() {

	if [ "$LADYBUG_CALL_METHOD" == "REST" ]; then
		resp=$(curl -sX POST ${c_URL_LADYBUG_NS}/clusters -H "${c_CT}" -d @- <<EOF
		{
			"name": "${v_CLUSTER_NAME}",
			"label": "",
			"description": "",
			"serviceType": "${v_SERVICE_TYPE}",
			"config": {
				"installMonAgent": "",
				"kubernetes": {
					"version": "1.23.14",
					"etcd": "local",
					"loadbalancer": "haproxy",
					"networkCni": "flannel",
					"podCidr": "10.244.0.0/16",
					"serviceCidr": "10.96.0.0/12",
					"serviceDnsDomain": "cluster.local"
				}
			},
			"controlPlane": [
				{
					"connection": "config-openstack-regionone",
					"count": 1,
					"spec": "ds2G",
					"rootDisk": {
						"type": "",
						"size": ""
					},
					"role": ""
				}
			],
			"worker": [
				{
					"connection": "config-openstack-regionone",
					"count": 1,
					"spec": "ds2G",
					"rootDisk": {
						"type": "",
						"size": ""
					},
					"role": ""
				}
			]
		}
EOF
		); 
		echo ${resp} | jq;

	elif [ "$LADYBUG_CALL_METHOD" == "GRPC" ]; then

		$APP_ROOT/src/grpc-api/cbadm/cbadm cluster create --config $APP_ROOT/src/grpc-api/cbadm/grpc_conf.yaml -i json -o json -d \
		'{
			"namespace":  "'${v_NAMESPACE}'",
			"ReqInfo": {
					"name": "'${v_CLUSTER_NAME}'",
					"label": "",
					"description": "",
					"serviceType": "'${v_SERVICE_TYPE}'",
					"config": {
						"installMonAgent": "no",
						"kubernetes": {
							"networkCni": "flannel",
							"podCidr": "10.244.0.0/16",
							"serviceCidr": "10.96.0.0/12",
							"serviceDnsDomain": "cluster.local",
							"loadbalancer": ""
						}
					},
					"controlPlane": [
						{
							"connection": "config-openstack-regionone",
							"count": 1,
							"spec": "ds2G",
							"rootDisk": {
								"type": "defalut",
								"size": "defalut"
							},
							"role": ""
						}
					],
					"worker": [
						{
							"connection": "config-openstack-regionone",
							"count": 2,
							"spec": "ds2G",
							"rootDisk": {
								"type": "defalut",
								"size": "defalut"
							},
							"role": ""
						}
					]
				}
		}'

	else
		echo "[ERROR] missing LADYBUG_CALL_METHOD"; exit -1;
	fi

}


# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then
	echo ""
	echo "------------------------------------------------------------------------------"
	create;
fi
