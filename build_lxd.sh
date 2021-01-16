#!/bin/bash
sudo apt-get -y update && sudo apt-get -y upgrade
sudo apt install -y gcc make liblxc1 liblxc-dev lxc-utils pkg-config
sudo snap install lxd
sudo usermod -aG lxd $USER

    sudo cat <<EOF | sudo lxd init --preseed
profiles:
  - name: default
    devices:
      root:
        path: /
        pool: default
        type: disk
      eth0:
        nictype: bridged
        parent: lxdbr0
        type: nic
networks:
  - name: lxdbr0
    type: bridge
    config:
      ipv4.address: auto
      ipv6.address: auto
storage_pools:
  - name: default
    driver: dir
    config:
      source: ""
EOF

newgrp lxd
