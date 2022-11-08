#!/bin/bash
IPS=""
for i in $@; do
IPS+="\""$i"\", "
done

sudo wget -q \
    https://storage.googleapis.com/kubernetes-the-hard-way/cfssl/1.4.1/linux/cfssl \
    https://storage.googleapis.com/kubernetes-the-hard-way/cfssl/1.4.1/linux/cfssljson
sudo chmod +x cfssl cfssljson
sudo mv cfssl cfssljson /usr/local/bin/
sudo mkdir -p /tmp/ca

sudo cat <<EOF | sudo tee ca-config.json
{
    "signing": {
        "default": {
            "expiry": "8760h"
        },
        "profiles": {
            "etcd": {
                "expiry": "8760h",
                "usages": ["signing","key encipherment","server auth","client auth"]
            }
        }
    }
}
EOF

sudo cat <<EOF | sudo tee ca-csr.json
{
  "CN": "etcd cluster",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "KR",
      "L": "SEOUL",
      "O": "Kubernetes",
      "OU": "ETCD-CA",
      "ST": "Cambridge"
    }
  ]
}
EOF

sudo cfssl gencert -initca ca-csr.json | sudo cfssljson -bare ca

sudo cat <<EOF | sudo tee etcd-csr.json
{
  "CN": "etcd",
  "hosts": [
    "localhost",
    "127.0.0.1",
    $IPS
    "kubernetes",
    "kubernetes.default",
    "kubernetes.default.svc",
    "kubernetes.default.svc.cluster",
    "kubernetes.default.svc.cluster.local"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "KR",
      "L": "SEOUL",
      "O": "Kubernetes",
      "OU": "etcd",
      "ST": "Cambridge"
    }
  ]
}
EOF

sudo cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=etcd etcd-csr.json | sudo cfssljson -bare etcd
sudo chmod 644 etcd-key.pem
sudo mv $HOME/ca.pem $HOME/etcd.pem $HOME/etcd-key.pem /tmp/ca
sudo rm -rf *.json *.csr *.pem