
Currently the GPU clients port should start from: 56001
The GPU scheduler port start from: 52001


deploy the $HOME/.kube/config content as a configmap:

docker push docker.io/kontonpuku666/fast-controller-manager:release 

kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config

