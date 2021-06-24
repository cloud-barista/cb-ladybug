#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$1" == "-h" ]; then 
	echo "./connectionconfig-add.sh [AWS/GCP/AZURE]"
	echo "./connectionconfig-add.sh GCP"
	exit 0
fi

# ------------------------------------------------------------------------------
# const

c_URL_SPIDER="http://localhost:1024/spider"
c_URL_TUMBLEBUG="http://localhost:1323/tumblebug"
c_CT="Content-Type: application/json"
c_AUTH="Authorization: Basic $(echo -n default:default | base64)"
c_AWS_DRIVER="aws-driver-v1.0"
c_GCP_DRIVER="gcp-driver-v1.0"
c_AZURE_DRIVER="azure-driver-v1.0"

# ------------------------------------------------------------------------------
# variables

# 1. CSP
if [ "$#" -gt 0 ]; then v_CSP="$1"; else	v_CSP="${CSP}"; fi
if [ "${v_CSP}" == "" ]; then 
	read -e -p "Cloud ? [AWS(default) or GCP or AZURE] : "  v_CSP
fi

if [ "${v_CSP}" == "" ]; then v_CSP="AWS"; fi
if [ "${v_CSP}" != "GCP" ] && [ "${v_CSP}" != "AWS" ] && [ "${v_CSP}" != "AZURE" ]; then echo "[ERROR] missing <cloud>"; exit -1;fi

v_CSP_LOWER="$(echo ${v_CSP} | tr [:upper:] [:lower:])"

# region
v_REGION="${REGION}"
if [ "${v_REGION}" == "" ]; then 
	read -e -p "region ? [예:asia-northeast3] : "  v_REGION
	if [ "${v_REGION}" == "" ]; then echo "[ERROR] missing region"; exit -1;fi
fi

if [ "${v_CSP}" == "AZURE" ]; then 

	# resource group
	v_RESOURCE_GROUP="${RESOURCE_GROUP}"
	if [ "${v_RESOURCE_GROUP}" == "" ]; then 
		read -e -p "resource group ? [예:cb-ladybugRG] : "  v_RESOURCE_GROUP
		if [ "${v_RESOURCE_GROUP}" == "" ]; then echo "[ERROR] missing resource group"; exit -1;fi
	fi

else

	# zone
	v_ZONE="${ZONE}"
	if [ "${v_ZONE}" == "" ]; then 
		read -e -p "zone ? [예:asia-northeast3-a] : "  v_ZONE
		if [ "${v_ZONE}" == "" ]; then v_ZONE="${v_REGION}-a";fi
	fi

fi


if [ "${v_CSP}" == "GCP" ]; then 
	v_DRIVER="${c_GCP_DRIVER}"
elif [ "${v_CSP}" == "AWS" ]; then 
	v_DRIVER="${c_AWS_DRIVER}"
else 
	v_DRIVER="${c_AZURE_DRIVER}"
fi

NM_CREDENTIAL="credential-${v_CSP_LOWER}"
NM_REGION="region-${v_CSP_LOWER}-${v_REGION}"
NM_CONFIG="config-${v_CSP_LOWER}-${v_REGION}"

# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Cloud                      is '${v_CSP}'"
echo "- Region                     is '${v_REGION}'"
if [ "${v_CSP}" == "AZURE" ]; then 
	echo "- Resource Group             is '${v_RESOURCE_GROUP}'"
else	
	echo "- Zone                       is '${v_ZONE}'"
fi
echo "- (Name of region)           is '${NM_REGION}'"
echo "- (Name of Connection Info.) is '${NM_CONFIG}'"


# ------------------------------------------------------------------------------
# Configuration Spider
init() {

	# region
	if [ "${v_CSP}" == "AZURE" ]; then
		curl -sX DELETE ${c_URL_SPIDER}/region/${NM_REGION} -H "${c_CT}" -o /dev/null -w "REGION.delete():%{http_code}\n"
		curl -sX POST   ${c_URL_SPIDER}/region              -H "${c_CT}" -o /dev/null -w "REGION.regist():%{http_code}\n" -d @- <<EOF
		{
		"ProviderName"     : "${v_CSP}", 
		"RegionName"       : "${NM_REGION}",
		"KeyValueInfoList" : [
			{"Key" : "location", "Value" : "${v_REGION}"},
			{"Key" : "ResourceGroup", "Value" : "${v_RESOURCE_GROUP}"}
		]
		}
EOF
	else	
		curl -sX DELETE ${c_URL_SPIDER}/region/${NM_REGION} -H "${c_CT}" -o /dev/null -w "REGION.delete():%{http_code}\n"
		curl -sX POST   ${c_URL_SPIDER}/region              -H "${c_CT}" -o /dev/null -w "REGION.regist():%{http_code}\n" -d @- <<EOF
		{
		"RegionName"       : "${NM_REGION}",
		"ProviderName"     : "${v_CSP}", 
		"KeyValueInfoList" : [
			{"Key" : "Region", "Value" : "${v_REGION}"},
			{"Key" : "Zone",   "Value" : "${v_ZONE}"}
		]
		}
EOF
	fi

	# config
	curl -sX DELETE ${c_URL_SPIDER}/connectionconfig/${NM_CONFIG} -H "${c_AUTH}" -H "${c_CT}" -o /dev/null -w "CONFIG.delete():%{http_code}\n"
	curl -sX POST   ${c_URL_SPIDER}/connectionconfig              -H "${c_AUTH}" -H "${c_CT}" -o /dev/null -w "CONFIG.regist():%{http_code}\n" -d @- <<EOF
	{
	"ConfigName"     : "${NM_CONFIG}",
	"ProviderName"   : "${v_CSP}", 
	"DriverName"     : "${v_DRIVER}", 
	"CredentialName" : "${NM_CREDENTIAL}", 
	"RegionName"     : "${NM_REGION}"
	}
EOF

}


# ------------------------------------------------------------------------------
# show init result
show() {
	echo "REGION";     curl -sX GET ${c_URL_SPIDER}/region/${NM_REGION}            -H "${c_AUTH}" -H "${c_CT}" | jq
	echo "CONFIG";     curl -sX GET ${c_URL_SPIDER}/connectionconfig/${NM_CONFIG}  -H "${c_AUTH}" -H "${c_CT}" | jq
}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	init;	show;
fi
