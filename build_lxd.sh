#!/bin/bash
sudo apt-get -y update && sudo apt-get -y upgrade
sudo snap install core
sudo snap install lxd

    cat <<EOF | lxd init --preseed
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
