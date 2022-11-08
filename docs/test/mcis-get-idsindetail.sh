#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then 
	echo "./mcis-get-idsindetail.sh <namespace> <mcis name> <vm name>"
	echo "./mcis-get-idsindetail.sh cb-mcks-ns cluster-01 vm-01"
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

# 3. vm 
if [ "$#" -gt 2 ]; then v_VM="$3"; fi
if [ "${v_VM}" == "" ]; then 
	read -e -p "Namespace ? : " v_VM
fi
if [ "${v_VM}" == "" ]; then echo "[ERROR] missing <vm name>"; exit -1; fi


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Namespace                  is '${v_NAMESPACE}'"
echo "- MCIS                       is '${v_MCIS}'"
echo "- VM                         is '${v_VM}'"

NM_TUMBLEBUG_NS="${c_URL_TUMBLEBUG}/ns/${v_NAMESPACE}"


# ------------------------------------------------------------------------------
# list
get_idsindetail() {
	curl -sX GET ${NM_TUMBLEBUG_NS}/mcis/${v_MCIS}/vm/${v_VM}?option=idsInDetail   -H "${c_AUTH}" -H "${c_CT}" | jq;
}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	get_idsindetail;
fi
