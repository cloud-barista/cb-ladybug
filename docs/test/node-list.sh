#!/bin/bash
# -----------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then 
	echo "./node-list.sh <namespace> <clsuter name>"
	echo "./node-list.sh cb-ladybug-ns cluster-01"
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


c_URL_LADYBUG_NS="${c_URL_LADYBUG}/ns/${v_NAMESPACE}"


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Namespace                  is '${v_NAMESPACE}'"
echo "- Cluster name               is '${v_CLUSTER_NAME}'"


# ------------------------------------------------------------------------------
# List Node
list() {

	if [ "$LADYBUG_CALL_METHOD" == "REST" ]; then
		
		curl -sX GET ${c_URL_LADYBUG_NS}/clusters/${v_CLUSTER_NAME}/nodes -H "${c_CT}" | jq;

	elif [ "$LADYBUG_CALL_METHOD" == "GRPC" ]; then

		$APP_ROOT/src/grpc-api/cbadm/cbadm node list --config $APP_ROOT/src/grpc-api/cbadm/grpc_conf.yaml -o json --ns ${v_NAMESPACE} --cluster ${v_CLUSTER_NAME}
		
	else
		echo "[ERROR] missing LADYBUG_CALL_METHOD"; exit -1;
	fi
	
}


# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	list;
fi
