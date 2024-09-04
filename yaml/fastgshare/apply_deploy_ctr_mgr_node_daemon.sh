#!/bin/bash
current_path=$(dirname "$0")

if [ ! -e /fastpod/library/libfast.so.1 ]; then
    if [ ! -e /fastpod/library ]; then
        sudo mkdir /fastpod/library
    fi
    sudo cp -r ${current_path}/install/libfast.so.1 /fastpod/library/
fi

if [ ! -e /models ]; then
    sudo mkdir /models
fi


kubectl apply -f ${current_path}/fastpod-controller-manager.yaml
sleep 10
kubectl apply -f ${current_path}/fastgshare-node-daemon.yaml