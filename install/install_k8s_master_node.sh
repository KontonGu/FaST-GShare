## Disable the swap
sudo swapoff -a
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab


cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

# cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
# net.ipv4.ip_forward = 1
# net.bridge.bridge-nf-call-ip6tables = 1
# net.bridge.bridge-nf-call-iptables = 1
# EOF
# sudo sysctl --system

cat <<EOF | sudo tee /etc/sysctl.d/kubernetes.conf
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system

#close firewall
sudo systemctl stop ufw
sudo systemctl disable ufw


# Installing Docker
sudo apt-get update

sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

sudo apt-get update 
# sudo apt-get install -y docker-ce=5:19.03.14~3-0~ubuntu-bionic \
# docker-ce-cli=5:19.03.14~3-0~ubuntu-bionic containerd.io containernetworking-plugins
# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update


VERSION_STRING=5:27.4.1-1~ubuntu.22.04~jammy
sudo apt-get install -y docker-ce=$VERSION_STRING docker-ce-cli=$VERSION_STRING containerd.io containernetworking-plugins

# sudo apt-get install -y docker-ce docker-ce-cli containerd.io containernetworking-plugins


# Installing kubeadm, kubelet and kubectl [version 1.26]
sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg # allow unprivileged APT programs to read this keyring
# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list   # helps tools such as command-not-found to work correctly

### can use apt-cache madison kubectl get check available versions
sudo apt-get update
sudo apt-get install -y kubelet=1.30.0-1.1 kubeadm=1.30.0-1.1 kubectl=1.30.0-1.1
sudo apt-mark hold kubelet kubeadm kubectl 

# sudo apt-get install -qy kubelet=1.24.2-00 kubectl=1.24.2-00 kubeadm=1.24.2-00
#sudo apt-mark hold kubelet kubeadm kubectl

# cd /etc/containerd/
# rm disable "CRI" from config.toml
sudo systemctl restart containerd

cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
sudo systemctl enable docker
sudo systemctl daemon-reload
sudo systemctl restart docker

sudo rm /etc/containerd/config.toml
sudo systemctl restart containerd

# sudo kubeadm init --pod-network-cidr=10.244.0.0/16 
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --cri-socket unix:///var/run/containerd/containerd.sock

mkdir -p $HOME/.kube
## [konton]: if you don't have the sudo permssion to copy the config file under /etc/kubernetes, execute followng command
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
# touch $HOME/.kube/config && sudo cat /etc/kubernetes/admin.conf > $HOME/.kube/config
# chown $(id -u):$(id -g) $HOME/.kube/config


# kubectl apply -f https://github.com/coreos/flannel/raw/master/Documentation/kube-flannel.yml --kubeconfig=$HOME/.kube/config
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml --kubeconfig=$HOME/.kube/config

sudo chown $USER:$USER /etc/kubernetes/admin.conf
sudo chmod 600 /etc/kubernetes/admin.conf

# remove sudo for docker container
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker

kubectl taint node `hostname` node-role.kubernetes.io/control-plane:NoSchedule-
kubectl taint node `hostname` node-role.kubernetes.io/master:NoSchedule-
kubectl label node `hostname` node-role.kubernetes.io/master=true