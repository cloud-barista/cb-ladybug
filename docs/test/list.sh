#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$1" == "-h" ]; then 
	echo "./list.sh <namespace> [all/config/ns/vpc/fw/ssh/image/spec/mcis]"
	echo "./list.sh cb-ladybug-ns ns"
	echo "./list.sh cb-ladybug-ns ns,config"
	exit 0
fi


# ------------------------------------------------------------------------------
# const

c_URL_SPIDER="http://localhost:1024/spider"
c_URL_TUMBLEBUG="http://localhost:1323/tumblebug"
c_CT="Content-Type: application/json"
c_AUTH="Authorization: Basic $(echo -n default:default | base64)"


# ------------------------------------------------------------------------------
# paramter

# 1. namespace
if [ "$#" -gt 0 ]; then v_NAMESPACE="$1"; else	v_NAMESPACE="${NAMESPACE}"; fi
if [ "${v_NAMESPACE}" == "" ]; then 
	read -e -p "Namespace ? : " v_NAMESPACE
fi
if [ "${v_NAMESPACE}" == "" ]; then echo "[ERROR] missing <namespace>"; exit -1; fi

# 2. query
if [ "$#" -gt 1 ]; then v_QUERY="$2"; fi

if [ "${v_QUERY}" == "" ]; then 
	read -e -p "Query ? [all/config/ns/vpc/fw/ssh/image/spec/mcis] : "  v_QUERY
fi
if [ "${v_QUERY}" == "" ]; then echo "[ERROR] missing <query>"; exit -1; fi
if [ "${v_QUERY}" == "all" ]; then v_QUERY="config,ns,vpc,fw,ssh,mcis"; fi


# variable - name
c_URL_TUMBLEBUG_NS="${c_URL_TUMBLEBUG}/ns/${v_NAMESPACE}"


# ------------------------------------------------------------------------------
# list
list() {
	if [[ "${v_QUERY}" == *"config"* ]]; then	echo "@_CONFIG_@";		curl -sX GET ${c_URL_SPIDER}/connectionconfig          			-H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"region"* ]]; then	echo "@_REGION_@";		curl -sX GET ${c_URL_SPIDER}/region          								-H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"ns"* ]]; then			echo "@_NAMESPACE_@";	curl -sX GET ${c_URL_TUMBLEBUG}/ns                       		-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"vpc"* ]]; then		echo "@_VPC_@";				curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/vNet						-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"fw"* ]]; then			echo "@_FW_@";				curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/securityGroup 	-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"ssh"* ]]; then		echo "@_SSH_@";				curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/sshKey					-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"image"* ]]; then	echo "@_IMAGE_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/image					-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"spec"* ]]; then		echo "@_SPEC_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/resources/spec						-H "${c_AUTH}" -H "${c_CT}" | jq; fi
	if [[ "${v_QUERY}" == *"mcis"* ]]; then		echo "@_MCIS_@";			curl -sX GET ${c_URL_TUMBLEBUG_NS}/mcis											-H "${c_AUTH}" -H "${c_CT}" | jq; fi
}


if [ "$1" != "-h" ]; then 
	list;
fi
