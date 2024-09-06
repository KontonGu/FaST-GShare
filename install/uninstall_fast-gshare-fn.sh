#!/bin/bash

current_dir=$(dirname "$0")
# clear all fastpods currently deployed
kubectl delete fastpods --all -n fast-gshare-fn

# helm uninstall the release fast-gshare
helm uninstall fast-gshare --namespace fast-gshare
kubectl delete pod -l fastgshare/role=dummyPod -n kube-system

kubectl delete -f ${current_dir}/mps_daemon.yaml