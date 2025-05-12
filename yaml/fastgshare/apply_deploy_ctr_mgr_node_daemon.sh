#!/bin/bash
current_dir=$(dirname "$0")
upper_dir=$(dirname "${current_dir}")
namespace_dir=$(dirname "${upper_dir}")
project_dir=$(dirname "${upper_dir}")

## create fast-gshare and fast-gshare-fn namespace if not existed
fastgshare_ns=$(kubectl get namespace fast-gshare --no-headers)
fastgshare_fn_ns=$(kubectl get namespace fast-gshare-fn --no-headers) 
if [ -z "${fastgshare_ns}" ]; then
    echo "creating fast-gshare and fast-gshare-fn namespace ..."
    kubectl apply -f ${namespace_dir}/namespace.yaml
fi

# check if the node has at least one NVIDIA GPU
if nvidia-smi &> /dev/null; then
    echo "NVIDIA GPU detected and nvidia-smi is working."
    kubectl label node `hostname` gpu=present
else
    echo "No NVIDIA GPU detected or nvidia-smi failed. exit."
    exit
fi

## check if the hook library is loaded to the diretory /fastpod/library/
if [ ! -e /fastpod/library/libhas.so.1 ]; then
    echo "fastpod hook library is missing. copy the file to the /fastpod/library..."
    if [ ! -e /fastpod/library ]; then
        sudo mkdir -p /fastpod/library
    fi
    sudo cp -r ${project_dir}/install/libhas.so.1 /fastpod/library/
fi

## check if the models dir is created
if [ ! -e /models ]; then
    echo "models dir is missing. creating /models."
    sudo mkdir /models
fi

## deploy the kube-config configmap if not existed
existed_config=$(kubectl get configmap kube-config -n kube-system --no-headers)
if [ -z "${existed_config}" ]; then
    echo "creating kube config configmap ..."
    kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config
fi


kubectl apply -f ${current_dir}/mps_daemon.yaml
sleep 3
kubectl apply -f ${current_dir}/fastpod-controller-manager.yaml
sleep 5
kubectl apply -f ${current_dir}/fastgshare-node-daemon.yaml