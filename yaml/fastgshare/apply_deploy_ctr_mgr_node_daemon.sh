#!/bin/bash
current_path=$(dirname "$0")

# check if the node has at least one NVIDIA GPU
if nvidia-smi &> /dev/null; then
    echo "NVIDIA GPU detected and nvidia-smi is working."
    kubectl label node `hostname` gpu=present
else
    echo "No NVIDIA GPU detected or nvidia-smi failed. exit."
    exit
fi

if [ ! -e /fastpod/library/libfast.so.1 ]; then
    echo "fastpod hook library is missing. copy the file to the /fastpod/library..."
    if [ ! -e /fastpod/library ]; then
        sudo mkdir /fastpod/library
    fi
    sudo cp -r ${project_dir}/install/libfast.so.1 /fastpod/library/
fi

if [ ! -e /models ]; then
    echo "models dir is missing. creating /models."
    sudo mkdir /models
fi

kubectl apply -f ${current_path}/mps_daemon.yaml
sleep 3
kubectl apply -f ${current_path}/fastpod-controller-manager.yaml
sleep 5
kubectl apply -f ${current_path}/fastgshare-node-daemon.yaml