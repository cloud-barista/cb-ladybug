#!/bin/bash
# kubeadm-config 정의
# - controlPlaneEndpoint 에 LB 지정 (9998 포트)
# - advertise-address 에 Public IP 지정
ALL="$@"
LIST=($ALL)
ETCDLIST=""
for ((i=5; i<${#LIST[@]}; i++)); do
if [ $i -eq 5 ]; then
ETCDLIST+="- https://${LIST[$i]}:2379"
elif [ $i -lt ${#LIST[@]} ]; then
ETCDLIST+=$'\n'"      - https://${LIST[$i]}:2379"
fi
done

cat << EOF > kubeadm-config.yaml
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
imageRepository: k8s.gcr.io
controlPlaneEndpoint: $4:$5
dns:
  type: CoreDNS
apiServer:
  extraArgs:
    advertise-address: $4
    authorization-mode: Node,RBAC
etcd:
  external:
    endpoints:
      $ETCDLIST
    caFile: /etc/kubernetes/pki/etcd/ca.pem
    certFile: /etc/kubernetes/pki/etcd/etcd.pem
    keyFile: /etc/kubernetes/pki/etcd/etcd-key.pem
networking:
  dnsDomain: $3
  podSubnet: $1
  serviceSubnet: $2
controllerManager: {}
scheduler: {}
EOF

# Control-plane init
sudo kubeadm init --v=5 --upload-certs --config kubeadm-config.yaml

# control-plane leader 의 경우
# - mcks-bootstrap 데몬이 자동 실행
#systemctl status mcks-bootstrap
