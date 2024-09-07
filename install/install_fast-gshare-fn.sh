##!/bin/bash
current_dir=$(dirname "$0")
project_dir=$(dirname "$current_dir")

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

## clear fastpod existed deployemnt configuration and use helm to intall fast-gshare-fn 
## which already includes fastpod deployment
existed_fastpods=$(kubectl get pods -n fast-gshare --no-headers)
if [ -n "$existed_fastpods" ]; then
    bash ${project_dir}/yaml/fastgshare/clean_deploy_ctr_mgr_node_daemon.sh
fi

echo "creating mps daemon ....."
kubectl apply -f ${current_dir}/mps_daemon.yaml

## install FaST-GShare-Function
kubectl apply -f ${project_dir}/namespace.yaml

## deploy the kube-config configmap if not existed
existed_config=$(kubectl get configmap kube-config -n kube-system --no-headers)
if [ -z "${existed_config}" ]; then
    echo "creating kube config configmap ..."
    kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config
fi
 
helm install fast-gshare ./chart/fastgshare --namespace fast-gshare --set functionNamespace=fast-gshare-fn  \
	--set  fastpodControllerManager.image="docker.io/kontonpuku666/fastpod-controller-manager:mps_test"


