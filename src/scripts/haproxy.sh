#!/bin/bash
sudo bash -c "cat << EOF > /etc/haproxy/haproxy.cfg
global
  log 127.0.0.1 local0
  maxconn 2000
  uid 0
  gid 0
  daemon
defaults
  log global
  mode tcp
  option dontlognull
  timeout connect 5000ms
  timeout client 50000ms
  timeout server 50000ms
frontend apiserver
  bind :9998
  default_backend apiserver
backend apiserver
  balance roundrobin
SERVERS
EOF"

# haproxy 재시작
sudo systemctl restart haproxy
