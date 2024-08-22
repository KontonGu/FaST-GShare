kubectl delete -f ./fastgshare-node-daemon.yaml
kubectl delete -f ./fastpod-controller-manager.yaml
kubectl delete pod -l fastgshare/role=dummyPod -n kube-system