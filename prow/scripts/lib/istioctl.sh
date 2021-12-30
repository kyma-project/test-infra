install_istioctl() {
    wget https://github.com/istio/istio/releases/download/1.11.3/istioctl-1.11.3-linux-amd64.tar.gz
    tar zxvf istioctl-1.11.3-linux-amd64.tar.gz -C /usr/local/bin/
    export ISTIOCTL_PATH=/usr/local/bin/istioctl
}