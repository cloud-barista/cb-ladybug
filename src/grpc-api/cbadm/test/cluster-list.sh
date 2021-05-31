#!/bin/bash
# -----------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then 
	echo "cluster-list.sh [GCP/AWS]"
	echo "    ./cluster-list.sh GCP"
	exit 0; 
fi


# ------------------------------------------------------------------------------
# const
c_URL_LADYBUG="http://localhost:8080/ladybug"
c_CT="Content-Type: application/json"


# -----------------------------------------------------------------
# parameter

# 1. CSP
if [ "$#" -gt 0 ]; then v_CSP="$1"; else	v_CSP="${CSP}"; fi
if [ "${v_CSP}" == "" ]; then 
	read -e -p "Cloud ? [AWS(default) or GCP] : "  v_CSP
fi
if [ "${v_CSP}" == "" ]; then v_CSP="AWS"; fi
if [ "${v_CSP}" != "GCP" ] && [ "${v_CSP}" != "AWS" ]; then echo "[ERROR] missing <cloud>"; exit -1;fi

# PREFIX
if [ "${v_CSP}" == "GCP" ]; then 
	v_PREFIX="cb-gcp"
else
	v_PREFIX="cb-aws"
fi

NM_NAMESPACE="${v_PREFIX}-namespace"
c_URL_LADYBUG_NS="${c_URL_LADYBUG}/ns/${NM_NAMESPACE}"


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Prefix                     is '${v_PREFIX}'"
echo "- Namespace                  is '${NM_NAMESPACE}'"


# ------------------------------------------------------------------------------
# List a cluster
list() {

	$APP_ROOT/src/grpc-api/cbadm/cbadm cluster list --config $APP_ROOT/src/grpc-api/cbadm/grpc_conf.yaml -o json --ns ${NM_NAMESPACE}

}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	list;
fi
