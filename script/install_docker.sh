#!/bin/bash

set -o nounset
#set -o errexit
#set -o xtrace

# default version, can be overridden by cmd line options
export DOCKER_VER=19.03.5
export REGISTRY_MIRROR=CN
export CONSOLE=${CONSOLE:-false}
export os_type=`cat /etc/os-release | grep "^ID=" | awk -F= '{print $2}' | tr -d [:punct:]`

function install_docker() {
    # check if a container runtime is already installed
    systemctl status docker | grep Active | grep -q running && sudo systemctl restart docker && {
        echo "[WARN] docker is already running."
        return 0
    }

    DOCKER_URL="https://rainbond-pkg.oss-cn-shanghai.aliyuncs.com/offline/docker/docker-${DOCKER_VER}.tgz"

    sudo mkdir -p /usr/local/bin /etc/docker /opt/docker/down
    if [[ -f "/opt/docker/down/docker-${DOCKER_VER}.tgz" ]]; then
        echo "[INFO] docker binaries already existed"
    else
        echo -e "[INFO] \033[33mdownloading docker binaries\033[0m $DOCKER_VER"
        if [[ -e /usr/bin/curl ]]; then
            curl -C- -O --retry 3 "$DOCKER_URL" || {
                echo "[ERROR] downloading docker failed"
                exit 1
            }
        else
            wget -c "$DOCKER_URL" || {
                echo "[ERROR] downloading docker failed"
                exit 1
            }
        fi
        sudo mv ./docker-${DOCKER_VER}.tgz /opt/docker/down
    fi

    sudo bash -c 'tar zxf /opt/docker/down/docker-*.tgz -C /opt/docker/down &&
        mv /opt/docker/down/docker/* /usr/local/bin &&
        ln -sf /usr/local/bin/docker /bin/docker && rm -rf /opt/docker/down'

    echo "[INFO] generate docker service file"
    sudo bash -c 'cat >/etc/systemd/system/docker.service <<EOF
[Unit]
Description=Docker Application Container Engine
Documentation=http://docs.docker.io
[Service]
OOMScoreAdjust=-1000
Environment="PATH=/usr/local/bin:/bin:/sbin:/usr/bin:/usr/sbin"
ExecStart=/usr/local/bin/dockerd
ExecStartPost=/sbin/iptables -I FORWARD -s 0.0.0.0/0 -j ACCEPT
ExecReload=/bin/kill -s HUP \$MAINPID
Restart=on-failure
RestartSec=5
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
Delegate=yes
KillMode=process
[Install]
WantedBy=multi-user.target
EOF'

    # configuration for dockerd
    echo "[INFO] generate docker config file"
    if [[ "$REGISTRY_MIRROR" == CN ]]; then
        echo "[INFO] prepare register mirror for $REGISTRY_MIRROR"
        sudo bash -c 'cat >/etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://dockerhub.azk8s.cn",
    "https://docker.mirrors.ustc.edu.cn",
    "http://hub-mirror.c.163.com"
  ],
  "max-concurrent-downloads": 10,
  "max-concurrent-uploads": 10,
  "log-driver": "json-file",
  "log-level": "warn",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
    },
  "data-root": "/var/lib/docker"
}
EOF'
    else
        echo "[INFO] standard config without registry mirrors"
        sudo 'cat >/etc/docker/daemon.json <<EOF
{
  "max-concurrent-downloads": 10,
  "log-driver": "json-file",
  "log-level": "warn",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
    },
  "data-root": "/var/lib/docker"
}
EOF'
    fi

    if [[ -e /etc/centos-release || -e /etc/redhat-release ]]; then
        echo "[INFO] turn off selinux in CentOS/Redhat"
        sudo setenforce 0
        sudo sed -ir "s/^SELINUX=.*/SELINUX=disable/" /etc/selinux/config
    fi

    echo "[INFO] enable and start docker"
    sudo systemctl enable docker
    sudo systemctl daemon-reload && sudo systemctl restart docker && sleep 8
}

function add_user_in_ubuntu() {
    sudo useradd --create-home -s /bin/bash -g docker "$1"
    echo $1:$2 | sudo chpasswd
    echo -e "[INFO] \033[32mUser $1 has been created \033[0m"
}

function add_user_in_redhat() {
    sudo adduser -g docker "$1"
    echo $1:$2 | sudo chpasswd
    echo -e "[INFO] \033[32mUser $1 has been created \033[0m"
}

function add_user() {
    user=${DOCKER_USER:-"docker"}
    pass=${DOCKER_PASS:-'rbd-123465!'}
    
    sudo groupadd --force docker

    if id -u $user >/dev/null 2>&1; then
        sudo gpasswd -a $user docker && echo -e "[INFO] \033[31mUser $user already exists \033[0m"
    else
        case $os_type in
        centos|redhat)
        add_user_in_redhat "$user" "$pass"
        ;;
        ubuntu|debian)
        add_user_in_ubuntu "$user" "$pass"
        ;;
        *)
        echo -e "[INFO] \033[31mDoes not support $os_type temporarily \033[0m" && exit 1
        ;;
        esac
    fi

    $CONSOLE || add_ssh_rsa "$user"
}

function disable_firewalld() {
    sudo systemctl stop firewalld >/dev/null 2>&1
    sudo systemctl disable firewalld >/dev/null 2>&1
}

function disable_swap() {
    sudo swapoff -a
    sudo sed -nr "s/(.*swap.*)/#\1/p" /etc/fstab
}

function add_ssh_rsa() {
    user_id=`id -u $1`
    if [ "$user_id" -eq "0" ]; then
        sudo ls -d /root/.ssh >/dev/null 2>&1
        if [ "$?" -gt "0" ]; then
            sudo mkdir /root/.ssh
        else
            sudo ls /root/.ssh/authorized_keys >/dev/null 2>&1 && sudo cp /root/.ssh/authorized_keys ./.authorized_keys_tmp && sudo chmod 666 ./.authorized_keys_tmp
        fi
        echo "$SSH_RSA" >> ./.authorized_keys_tmp && mv ./.authorized_keys_tmp /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys && chown root:root /root/.ssh/authorized_keys && echo -e "[INFO] \033[32mSSH_RSA has been written \033[0m"
    else
        sudo ls -d /home/$1/.ssh >/dev/null 2>&1
        if [ "$?" -gt "0" ]; then
            sudo mkdir /home/$1/.ssh
        else
            sudo ls /home/$1/.ssh/authorized_keys >/dev/null 2>&1 && sudo cp /home/$1/.ssh/authorized_keys ./.authorized_keys_tmp && sudo chmod 666 ./.authorized_keys_tmp
        fi
        echo "$SSH_RSA" >> ./.authorized_keys_tmp && sudo mv ./.authorized_keys_tmp /home/$1/.ssh/authorized_keys && sudo chmod 600 /home/$1/.ssh/authorized_keys && sudo chown $1 /home/$1/.ssh/authorized_keys && echo -e "[INFO] \033[32mSSH_RSA has been written \033[0m"
    fi
}

function time_sync() {

    sudo ntpdate time.pool.aliyun.com >/dev/null 2>&1 || case $os_type in
    centos|redhat)
	sudo yum install -y --quiet ntpdate ;
	;;
    ubuntu|debian)
	sudo apt -q update && sudo apt install -q -y ntpdate;
	;;
    *)
    echo -e "[INFO] \033[31mDoes not support $os_type temporarily \033[0m" && exit 1
    ;;
    esac

    sudo ntpdate time.pool.aliyun.com && echo -e "[INFO] \033[32mTime synchronization is complete \033[0m"
}

function check_cpu(){
    local cpu=$(lscpu | awk '/^CPU\(/{print $2}')
    if [ ${cpu} -lt 2 ]; then
        echo -e "[WARN] \033[31mThe cpu is recommended to be at least 2C \033[0m"
    fi
}

function check_mem(){
    local mem=$(free -g | awk '/^Mem/{print $2}')
    if [ ${mem} -lt 3 ]; then
        echo -e "[WARN] \033[31mThe Memory is recommended to be at least 4G \033[0m"
    fi
}

function check_os() {
    os_version=`cat /etc/os-release | grep VERSION_ID|awk -F= '{print $2}'|tr -d [:punct:]`
    case $os_type in
    centos|redhat)
    if [ "$os_version" -lt "7" ]; then
        echo -e "[ERROR] \033[31mCentos or redhat version must be higher than 7.0  \033[0m" && exit 1
    fi
    ;;
    ubuntu)
    if [ "$os_version" -lt "1604" ]; then
        echo -e "[ERROR] \033[31mUbuntu version must be higher than 16.04  \033[0m" && exit 1
    fi
    ;;
    debian)
    if [ "$os_version" -lt "9" ]; then
        echo -e "[ERROR] \033[31mUbuntu version must be higher than 9.0  \033[0m" && exit 1
    fi
    ;;
    *)
    echo -e "[INFO] \033[31mDoes not support $os_type temporarily \033[0m" && exit 1
    ;;
    esac
}

function check_kernel() {
    kernel_version=`uname -r | awk -F. '{print $1}'`
    if [ "$kernel_version" -lt "4" ]; then 
        echo -e "[INFO] \033[31mKernel version must be higher than 4.0;Please upgrade the kernel to 4.0+ as soon as possible \033[0m"
    fi
}

# check os
check_os

# check kernel
check_kernel

# time sync
time_sync

# check cpu
check_cpu

# check mem
check_mem

# disable swap
disable_swap

# disable firewalld
disable_firewalld

# add docker user
add_user

# install docker
install_docker