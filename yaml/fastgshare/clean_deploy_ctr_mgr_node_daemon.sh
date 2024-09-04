#!/bin/bash
current_path=$(dirname "$0")
kubectl delete fastpods --all -n fast-gshare
kubectl delete -f ${current_path}/fastgshare-node-daemon.yaml
kubectl delete -f ${current_path}/fastpod-controller-manager.yaml
kubectl delete pod -l fastgshare/role=dummyPod -n kube-system