#!/bin/bash
ETCD_VER="v3.5.1"
ETCD_NAME=$(hostname -s)
NODE_IP=$(hostname -I)
NODE_IPS=(${NODE_IP})
regexp='([0-9]{1,3}\.){3}[0-9]{1,3}'
CLUSTER=""

for i in $@; do
if [[ "$i" =~ $regexp ]]; then
CLUSTER+=$i":2380,"
else
CLUSTER+=$i"=https://"
fi
done

sudo mkdir -p /etc/kubernetes/pki/etcd
sudo chown root.root $HOME/*.pem
sudo mv $HOME/ca.pem $HOME/etcd.pem $HOME/etcd-key.pem /etc/kubernetes/pki/etcd

sudo wget -q "https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz"
sudo tar zxf etcd-${ETCD_VER}-linux-amd64.tar.gz
sudo mv etcd-${ETCD_VER}-linux-amd64/etcd* /usr/local/bin/
sudo rm -rf etcd*

sudo cat <<EOF | sudo tee /etc/systemd/system/etcd.service
[Unit]
Description=etcd

[Service]
Type=exec
ExecStart=/usr/local/bin/etcd \\
  --name ${ETCD_NAME} \\
  --cert-file=/etc/kubernetes/pki/etcd/etcd.pem \\
  --key-file=/etc/kubernetes/pki/etcd/etcd-key.pem \\
  --peer-cert-file=/etc/kubernetes/pki/etcd/etcd.pem \\
  --peer-key-file=/etc/kubernetes/pki/etcd/etcd-key.pem \\
  --trusted-ca-file=/etc/kubernetes/pki/etcd/ca.pem \\
  --peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.pem \\
  --peer-client-cert-auth \\
  --client-cert-auth \\
  --initial-advertise-peer-urls https://${NODE_IPS[0]}:2380 \\
  --listen-peer-urls https://0.0.0.0:2380 \\
  --advertise-client-urls https://${NODE_IPS[0]}:2379 \\
  --listen-client-urls https://0.0.0.0:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster ${CLUSTER%\,} \\
  --initial-cluster-state new
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now etcd