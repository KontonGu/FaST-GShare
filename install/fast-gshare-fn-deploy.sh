##!/bin/bash
current_dir=$(dirname "$0")
project_dir=$(dirname "$current_dir")


# clear fastpod deployemnt configuration and use helm to intall fast-gshare-fn
bash ${project_dir}/yaml/fastgshare/clean_deploy_ctr_mgr_node_daemon.sh

## install FaST-GShare-Function
kubectl apply -f ${project_dir}/namespace.yaml
kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config 
make helm_install_fast-gshare-fn


