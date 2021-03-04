# kubeadm init
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address=$(dig +short myip.opendns.com @resolver1.opendns.com) --token-ttl=0

# install addon mode
sudo kubectl apply -f https://raw.githubusercontent.com/squat/kilo/master/manifests/kilo-kubeadm-flannel.yaml --kubeconfig=/etc/kubernetes/admin.conf

# apply flannel cni
sudo kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml --kubeconfig=/etc/kubernetes/admin.conf
