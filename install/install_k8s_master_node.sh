## Disable the swap
sudo swapoff -a
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab


cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
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

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

sudo apt-get update 
# sudo apt-get install -y docker-ce=5:19.03.14~3-0~ubuntu-bionic \
# docker-ce-cli=5:19.03.14~3-0~ubuntu-bionic containerd.io containernetworking-plugins

sudo apt-get install -y docker-ce docker-ce-cli containerd.io containernetworking-plugins

# Installing kubeadm, kubelet and kubectl [version 1.26]
sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl

sudo mkdir -p -m 755 /etc/apt/keyrings

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.26/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.26/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet=1.26.2-1.1 kubeadm=1.26.2-1.1 kubectl=1.26.2-1.1
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


kubectl apply -f https://github.com/coreos/flannel/raw/master/Documentation/kube-flannel.yml --kubeconfig=$HOME/.kube/config


# remove sudo for docker container
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker

kubectl taint node `hostname` node-role.kubernetes.io/control-plane:NoSchedule-
kubectl taint node `hostname` node-role.kubernetes.io/master:NoSchedule-
kubectl label node `hostname` node-role.kubernetes.io/master=true