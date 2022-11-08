#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then 
	echo "./mcis-delete.sh <namespace> <mcis name>"
	echo "./mcis-delete.sh cb-mcks-ns cluster-01"
	exit 0
fi

source ./conf.env

# ------------------------------------------------------------------------------
# const


# ------------------------------------------------------------------------------
# variables

# 1. namespace
if [ "$#" -gt 0 ]; then v_NAMESPACE="$1"; fi
if [ "${v_NAMESPACE}" == "" ]; then 
	read -e -p "Namespace ? : " v_NAMESPACE
fi
if [ "${v_NAMESPACE}" == "" ]; then echo "[ERROR] missing <namespace>"; exit -1; fi

# 2. mcis 
if [ "$#" -gt 1 ]; then v_MCIS="$2"; fi
if [ "${v_MCIS}" == "" ]; then 
	read -e -p "Namespace ? : " v_MCIS
fi
if [ "${v_MCIS}" == "" ]; then echo "[ERROR] missing <mcis name>"; exit -1; fi


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Namespace                  is '${v_NAMESPACE}'"
echo "- MCIS                       is '${v_MCIS}'"

NM_TUMBLEBUG_NS="${c_URL_TUMBLEBUG}/ns/${v_NAMESPACE}"


# ------------------------------------------------------------------------------
# list
delete() {
	curl -sX DELETE ${NM_TUMBLEBUG_NS}/mcis/${v_MCIS}?option=force -H "${c_AUTH}" -H "${c_CT}" | jq;
}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	delete;	
fi
