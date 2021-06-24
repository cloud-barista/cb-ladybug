#!/bin/bash
# ------------------------------------------------------------------------------
# usage
if [ "$1" == "-h" ]; then 
	echo "./init.ns.sh <namespace>"
	echo "./init.ns.sh cb-ladybug-ns "
	exit 0
fi

# ------------------------------------------------------------------------------
# const

c_URL_TUMBLEBUG="http://localhost:1323/tumblebug"
c_CT="Content-Type: application/json"
c_AUTH="Authorization: Basic $(echo -n default:default | base64)"

# ------------------------------------------------------------------------------
# variables

v_NAMESPACE="$1"
if [ "${v_NAMESPACE}" == "" ]; then read -e -p "namespace ? : "  v_NAMESPACE;	fi
if [ "${v_NAMESPACE}" == "" ]; then echo "[ERROR] missing <namespace>"; exit -1; fi


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- (Name of namespace)        is '${v_NAMESPACE}'"


# ------------------------------------------------------------------------------
# Configuration Thumblebug
init() {

	# namespace
	curl -sX DELETE ${c_URL_TUMBLEBUG}/ns -H "${c_AUTH}" -H "${c_CT}" -o /dev/null -w "NAMESPACE.delete():%{http_code}\n"
	curl -sX POST   ${c_URL_TUMBLEBUG}/ns -H "${c_AUTH}" -H "${c_CT}" -o /dev/null -w "NAMESPACE.regist():%{http_code}\n" -d @- <<EOF
	{
	"name"        : "${v_NAMESPACE}",
	"description" : ""
	}
EOF

}


# ------------------------------------------------------------------------------
# show init result
show() {
	echo "NAMESPACE";  curl -sX GET ${c_URL_TUMBLEBUG}/ns/${v_NAMESPACE}          -H "${c_AUTH}" -H "${c_CT}" | jq
}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	init;	show;
fi
