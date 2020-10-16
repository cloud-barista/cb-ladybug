#!/bin/bash
# -----------------------------------------------------------------
# usage
if [ "$#" -lt 1 ]; then 
	echo "ladybug.sh [create/destroy] [GCP/AWS] <clsuter name> <spec> <worker node count>"
	echo "    ./ladybug.sh create GCP cb-cluster n1-standard-2 2"
	echo "    ./ladybug.sh create AWS cb-cluster t2.medium 2"
	echo "    ./ladybug.sh destroy GCP cb-cluster"
	exit 0; 
fi


# ------------------------------------------------------------------------------
# const
c_URL_LADYBUG="http://localhost:8080/ladybug"
c_CT="Content-Type: application/json"


# -----------------------------------------------------------------
# parameter

# 1. METHOD
if [ "$#" -gt 0 ]; then v_METHOD="$1"; else	v_METHOD="${METHOD}"; fi
if [ "${v_METHOD}" == "" ]; then 
	read -e -p "Method [create.destroy/list/get] ? : "  v_METHOD
fi
if [ "${v_METHOD}" == "" ]; then echo "[ERROR] missing <method>"; exit -1; fi


# 2. CSP
if [ "$#" -gt 1 ]; then v_CSP="$2"; else	v_CSP="${CSP}"; fi
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
# # 2. PREFIX
# if [ "$#" -gt 0 ]; then v_PREFIX="$2"; else	v_PREFIX="${PREFIX}"; fi
# if [ "${v_PREFIX}" == "" ]; then 
# 	read -e -p "Name prefix ? : "  v_PREFIX
# fi
# if [ "${v_PREFIX}" == "" ]; then v_PREFIX="${v_CSP}"; fi

# 3. Cluster Name
if [ "$#" -gt 2 ]; then v_CLUSTER_NAME="$3"; else	v_METHOD="${CLUSTER_NAME}"; fi
if [ "${v_CLUSTER_NAME}" == "" ]; then 
	read -e -p "Cluster name  ? : "  v_CLUSTER_NAME
fi
if [ "${v_CLUSTER_NAME}" == "" ]; then echo "[ERROR] missing <cluster name>"; exit -1; fi

# 4~5 "create" 인 경우만
if [ "${v_METHOD}" == "create" ]; then
	# 4. SPEC
	if [ "$#" -gt 3 ]; then v_SPEC="$4"; else	v_SPEC="${SPEC}"; fi
	if [ "${v_SPEC}" == "" ]; then 
		read -e -p "spec ? [예:n1-standard-2, t2.medium] : "  v_SPEC
	fi
	if [ "${v_CSP}" == "" ]; then 
		if [ "${v_CSP}" == "GCP" ]; then 
			v_SPEC="n1-standard-2"
		else
			v_SPEC="t2.medium"
		fi
	fi

	# 4. WORKER_NODE_COUNT
	if [ "$#" -gt 4 ]; then v_WORKER_NODE_COUNT="$5"; else	v_WORKER_NODE_COUNT="${WORKER_NODE_COUNT}"; fi
	if [ "${v_WORKER_NODE_COUNT}" == "" ]; then 
		read -e -p "worker node count [예:2] : "  v_WORKER_NODE_COUNT
	fi
	if [ "${v_WORKER_NODE_COUNT}" == "" ]; then v_WORKER_NODE_COUNT="2"; fi

fi

NM_NAMESPACE="${v_PREFIX}-namespace"
NM_CONFIG="${v_PREFIX}-config"
c_URL_LADYBUG_NS="${c_URL_LADYBUG}/ns/${NM_NAMESPACE}"


# ------------------------------------------------------------------------------
# print info.
echo ""
echo "[INFO]"
echo "- Method                     is '${v_METHOD}'"
echo "- Prefix                     is '${v_PREFIX}'"
echo "- Cuseter name               is '${v_CLUSTER_NAME}'"
echo "- Spec                       is '${v_SPEC}'"
echo "- Worker node count          is '${v_WORKER_NODE_COUNT}'"
echo "- Namespace                  is '${NM_NAMESPACE}'"
echo "- (Name of Connection Info.) is '${NM_CONFIG}'"


# ------------------------------------------------------------------------------
# list
list() {

	curl -sX GET ${c_URL_LADYBUG_NS}/clusters  -H "${c_CT}" | jq
}

# ------------------------------------------------------------------------------
# get Infrastructure
get() {
	curl -sX GET ${c_URL_LADYBUG_NS}/clusters/${v_CLUSTER_NAME}    -H "${c_CT}" | jq;
}

# ------------------------------------------------------------------------------
# Create Infrastructure
create() {

	rm -f kube-config.yaml

	resp=$(curl -sX POST ${c_URL_LADYBUG_NS}/clusters -H "${c_CT}" -d @- <<EOF
	{
		"name"                     : "${v_CLUSTER_NAME}",
		"control-plane-node-count" : 1,
		"control-plane-node-spec"  : "${v_SPEC}",
		"worker-node-count"        : ${v_WORKER_NODE_COUNT},
		"worker-node-spec"         : "${v_SPEC}" 
	}
EOF
	); echo ${resp} | jq -r ".\"cluster-config\"" > kube-config.yaml
}

# ------------------------------------------------------------------------------
# Destroy Infrastructure
destroy() {

	curl -sX DELETE ${c_URL_LADYBUG_NS}/clusters/${v_CLUSTER_NAME}    -H "${c_CT}" | jq;

}

# ------------------------------------------------------------------------------
if [ "$1" != "-h" ]; then 
	echo ""
	echo "------------------------------------------------------------------------------"
	if [ "${v_METHOD}" == "list" ];		then	list; fi
	if [ "${v_METHOD}" == "get" ];		then	get; fi
	if [ "${v_METHOD}" == "create" ];	then	create;	fi
	if [ "${v_METHOD}" == "destroy" ];	then	destroy; fi
fi
