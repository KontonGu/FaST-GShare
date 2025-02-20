#!/bin/bash

ubuntu_version=$(lsb_release -sr)
echo ${ubuntu_version}
    
if [[ "${ubuntu_version}" == "20.04" ]]; then
    sudo apt-get update
    wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-ubuntu2004.pin
    sudo mv cuda-ubuntu2004.pin /etc/apt/preferences.d/cuda-repository-pin-600
    wget https://developer.download.nvidia.com/compute/cuda/12.2.0/local_installers/cuda-repo-ubuntu2004-12-2-local_12.2.0-535.54.03-1_amd64.deb
    sudo dpkg -i cuda-repo-ubuntu2004-12-2-local_12.2.0-535.54.03-1_amd64.deb
    sudo cp /var/cuda-repo-ubuntu2004-12-2-local/cuda-*-keyring.gpg /usr/share/keyrings/
    sudo apt-get update
    sudo apt-get -y install nvidia-driver-535
    sudo apt-get -y install cuda
    sudo apt-mark hold nvidia-driver-535
    sudo apt-mark hold cuda
    
elif [[ "${ubuntu_version}" == "22.04" ]]; then
    wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-ubuntu2204.pin
    sudo mv cuda-ubuntu2204.pin /etc/apt/preferences.d/cuda-repository-pin-600
    wget https://developer.download.nvidia.com/compute/cuda/12.2.0/local_installers/cuda-repo-ubuntu2204-12-2-local_12.2.0-535.54.03-1_amd64.deb
    sudo dpkg -i cuda-repo-ubuntu2204-12-2-local_12.2.0-535.54.03-1_amd64.deb
    sudo cp /var/cuda-repo-ubuntu2204-12-2-local/cuda-*-keyring.gpg /usr/share/keyrings/
    sudo apt-get update
    sudo apt-get -y install cuda
    sudo apt-mark hold cuda
fi

sudo apt-mark hold nvidia-driver-535

CUDA_PATH=/usr/local/cuda
CUDA_PATH_LINE="export PATH=\$CUDA_PATH/bin:\$PATH"
CUDA_LD_LIBRARY_PATH_LINE="export LD_LIBRARY_PATH=\$CUDA_PATH/lib64:\$LD_LIBRARY_PATH"
echo "CUDA_PATH=${CUDA_PATH}" >> ~/.bashrc
echo "$CUDA_PATH_LINE" >> ~/.bashrc
echo "$CUDA_LD_LIBRARY_PATH_LINE" >> ~/.bashrc
source ~/.bashrc
echo "CUDA paths have been added to .bashrc and changes have been applied."

## New version
# sudo apt-get update
# sudo apt-get install -y cuda-drivers   ## install cuda driver with latest version 
# sudo apt-get update
# sudo apt-get -y install cuda-toolkit-12-6