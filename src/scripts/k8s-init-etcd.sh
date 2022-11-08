#!/bin/bash

POD_CIDR="$1"
SERVICE_CIDR="$2"
SERVICE_DNS_DOMAIN="$3"
PUBLIC_IP="$4"
PRIVATE_IP="$5"
PORT="$6"
SERVICE_TYPE="$7"

# kubeadm-config 정의
# - controlPlaneEndpoint 에 LB 지정 (9998 포트)
# - advertise-address 에 Multi Cloud Type인 경우 Public IP 지정, Single Cloud Type인 경우 Private IP 지정

ADVERTISE_ADDR=${PUBLIC_IP}
if [ "${SERVICE_TYPE}" == "single" ]; then
    ADVERTISE_ADDR=${PRIVATE_IP}
fi

ALL="$@"
LIST=($ALL)
ETCDLIST=""
for ((i=7; i<${#LIST[@]}; i++)); do
if [ $i -eq 7 ]; then
ETCDLIST+="- https://${LIST[$i]}:2379"
elif [ $i -lt ${#LIST[@]} ]; then
ETCDLIST+=$'\n'"      - https://${LIST[$i]}:2379"
fi
done

cat << EOF > kubeadm-config.yaml
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
imageRepository: k8s.gcr.io
controlPlaneEndpoint: ${PUBLIC_IP}:${PORT}
dns:
  type: CoreDNS
apiServer:
  extraArgs:
    advertise-address: ${ADVERTISE_ADDR}
    authorization-mode: Node,RBAC
  certSANs:
  - ${PUBLIC_IP}
  - ${PRIVATE_IP}
etcd:
  external:
    endpoints:
      $ETCDLIST
    caFile: /etc/kubernetes/pki/etcd/ca.pem
    certFile: /etc/kubernetes/pki/etcd/etcd.pem
    keyFile: /etc/kubernetes/pki/etcd/etcd-key.pem
networking:
  dnsDomain: ${SERVICE_DNS_DOMAIN}
  podSubnet: ${POD_CIDR}
  serviceSubnet: ${SERVICE_CIDR}
controllerManager: {}
scheduler: {}
EOF

if [ "${SERVICE_TYPE}" == "single" ]; then

cat << EOF >> kubeadm-config.yaml
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: external
EOF

fi


# Control-plane init
sudo kubeadm init --v=5 --upload-certs --config kubeadm-config.yaml

# control-plane leader 의 경우
# - mcks-bootstrap 데몬이 자동 실행
#systemctl status mcks-bootstrap
