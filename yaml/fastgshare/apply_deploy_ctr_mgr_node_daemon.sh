#!/bin/bash
current_path=$(dirname "$0")
kubectl apply -f ${current_path}/fastpod-controller-manager.yaml
sleep 10
kubectl apply -f ${current_path}/fastgshare-node-daemon.yaml