#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$1" == "-h" ]; then 
	echo "./get.sh [GCP/AWS] [config/region/ns/vpc/fw/ssh/image/spec/mcis/ip] <cluster name> <spec name>"
	echo "./get.sh GCP ns"
	echo "./get.sh GCP config,region,ns"
  echo "./get.sh GCP mcis cb-cluster"
  echo "./get.sh GCP spec cb-cluster n1-standard-2"
	exit 0
fi


# ------------------------------------------------------------------------------
# const

c_URL_SPIDER="http://localhost:1024/spider"
c_URL_TUMBLEBUG="http://localhost:1323/tumblebug"
c_CT="Content-Type: application/json"
c_AUTH="Authorization: Basic $(echo -n default:default | base64)"

# ------------------------------------------------------------------------------
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
  v_CSP_LOWER="gcp"
else
	v_PREFIX="cb-aws"
  v_CSP_LOWER="aws"
fi

# # PREFIX
# if [ "$#" -gt 0 ]; then v_PREFIX="$1"; else	v_PREFIX="${PREFIX}"; fi

# if [ "${v_PREFIX}" == "" ]; then 
# 	read -e -p "Name prefix ? : "  v_PREFIX
# fi
# if [ "${v_PREFIX}" == "" ]; then echo "[ERROR] missing <prefix>"; exit -1; fi

# query
if [ "$#" -gt 1 ]; then v_QUERY="$2"; fi

if [ "${v_QUERY}" == "" ]; then 
	read -e -p "Query ? [all/ns/vpc/fw/ssh/image/spec/mcis/ip] : "  v_QUERY
fi
if [ "${v_QUERY}" == "" ]; then echo "[ERROR] missing <query>"; exit -1; fi
if [ "${v_QUERY}" == "all" ]; then v_QUERY="config/region/ns/vpc/fm/ssh/image/spec/mcis/ip"; fi

# Cluster Name
if [ "$#" -gt 2 ]; then v_CLUSTER_NAME="$3"; else	v_CLUSTER_NAME="${CLUSTER_NAME}"; fi
if [ "${v_CLUSTER_NAME}" == "" ]; then 
	read -e -p "Cluster name  ? : "  v_CLUSTER_NAME
fi
if [ "${v_CLUSTER_NAME}" == "" ]; then echo "[ERROR] missing <cluster name>"; exit -1; fi

# Spec Name
if [ "$#" -gt 3 ]; then v_SPEC="$4"; else	v_SPEC="${SPEC}"; fi
if [ "${v_SPEC}" == "" ]; then 
	read -e -p "Spec name  ? : "  v_SPEC
fi
if [ "${v_SPEC}" == "" ]; then echo "[ERROR] missing <spec name>"; exit -1; fi

c_CREDENTIAL=${v_CLUSTER_NAME}.pem
c_USERNAME="ubuntu"

# variable - name
NM_NAMESPACE="${v_PREFIX}-namespace"
NM_CONFIG="${v_PREFIX}-config"
NM_VPC="${v_CLUSTER_NAME}-${v_CSP_LOWER}-vpc"
NM_FW="${v_CLUSTER_NAME}-${v_CSP_LOWER}-allow-external"
NM_SSH_KEY="${v_CLUSTER_NAME}-${v_CSP_LOWER}-sshkey"
NM_REGION="${v_PREFIX}-region"
NM_IMAGE="${NM_CONFIG}-${REGION}-Ubuntu1804"
NM_MACHINE="${v_SPEC//./-}-spec"
NM_MCIS="${v_CLUSTER_NAME}" 

c_URL_TUMBLEBUG_NS="${c_URL_TUMBLEBUG}/ns/${NM_NAMESPACE}"

# ------------------------------------------------------------------------------
# print info.
echo "[INFO]"
echo "- Prefix                     is '${v_PREFIX}'"
echo "- Namespace                  is '${NM_NAMESPACE}'"
echo "- (Name of Connection Info.) is '${NM_CONFIG}'"
echo "- (Name of Region)           is '${NM_CONFIG}'"
echo "- (Name of vpc)              is '${NM_VPC}'"
echo "- (Name of firewall)         is '${NM_FW}'"
echo "- (Name of ssh key)          is '${NM_SSH_KEY}'"
echo "- (Name of image)            is '${NM_IMAGE}'"
echo "- (Name of spec)             is '${NM_MACHINE}'"
echo "- (Name of MCIS)             is '${NM_MCIS}'"


# ------------------------------------------------------------------------------
# get Infrastructure
get() {
	if [[ "${v_QUERY}" == *"config"* ]]; then		echo "@_CONFIG_@";		curl -sX GET ${c_URL_SPIDER}/connectionconfig/${NM_CONFIG}          -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"region"* ]]; then		echo "@_REGION_@";		curl -sX GET ${c_URL_SPIDER}/region/${NM_REGION}                 	-H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"ns"* ]]; then			echo "@_NAMESPACE_@";	curl -sX GET ${c_URL_TUMBLEBUG}/ns/${NM_NAMESPACE}                  -H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"vpc"* ]]; then			echo "@_VPC_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/vNet/${NM_VPC}         -H "${c_AUTH}" -H "${c_CT}" -d '{"connectionName" : "'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"vpc.spider"* ]]; then	echo "@_VPC_SPIDER@";	curl -sX GET ${c_URL_SPIDER}/vpc/${NM_VPC}                          -H "${c_AUTH}" -H "${c_CT}" -d '{"ConnectionName":"'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"fw"* ]]; then			echo "@_FW_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/securityGroup/${NM_FW} -H "${c_AUTH}" -H "${c_CT}" -d '{"connectionName" : "'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"ssh"* ]]; then			echo "@_SSH_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/sshKey/${NM_SSH_KEY}   -H "${c_AUTH}" -H "${c_CT}" -d '{"connectionName" : "'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"image"* ]]; then		echo "@_IMAGE_@";		curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/image/${NM_IMAGE}      -H "${c_AUTH}" -H "${c_CT}" -d '{"connectionName" : "'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"spec"* ]]; then			echo "@_SPEC_@";		curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/spec/${NM_MACHINE}     -H "${c_AUTH}" -H "${c_CT}" -d '{"connectionName" : "'${NM_CONFIG}'"}' | jq; fi
	if [[ "${v_QUERY}" == *"mcis"* ]]; then			echo "@_MCIS_@";		curl -sX GET ${c_URL_TUMBLEBUG_NS}/mcis/${NM_MCIS}                  -H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"ip"* ]]
	then
		RESP=$(curl -sX GET ${c_URL_TUMBLEBUG_NS}/mcis/${NM_MCIS} -H "${c_AUTH}" -H "${c_CT}")
		echo ${RESP}| jq -r ".vm | .[0].publicIP"
		echo "ssh -i $(pwd)/${c_CREDENTIAL} ${c_USERNAME}@$(echo ${RESP}| jq -r ".vm | .[0].publicIP")"
	fi
}


# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	get;
fi
