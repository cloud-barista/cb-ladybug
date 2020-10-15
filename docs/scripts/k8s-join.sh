#!/bin/bash
# direct run : k8s join
END_POINT=$1
TOKEN=$2
HASH=$3
sudo kubeadm join $END_POINT --token $TOKEN --discovery-token-ca-cert-hash sha256:$HASH