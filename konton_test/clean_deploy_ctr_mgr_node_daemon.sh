kubectl delete -f ../yaml/fastgshare/fastgshare-node-daemon.yaml
kubectl delete -f ../yaml/fastgshare/fastpod-controller-manager.yaml
kubectl delete pod -l fastgshare/role=dummyPod -n kube-system