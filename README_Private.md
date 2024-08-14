
Currently the GPU clients port should start from: 56001
The GPU scheduler port start from: 52001


deploy the $HOME/.kube/config content as a configmap:

docker push docker.io/kontonpuku666/fastpod-controller-manager:release 

docker push docker.io/kontonpuku666/fast-configurator:release 

kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config


clean specific ctr images:
sudo ctr -n k8s.io i rm docker.io/kontonpuku666/fast-configurator:release

apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fastgshare-node-daemon
  namespace: kube-system
  labels:
    fastgshare: configurator-scheduler-node-daemon
spec:
  selector:
    matchLabels:
        fastgsahre: configurator-scheduler-node-daemon
  template:
    metadata:
      labels:
        fastgsahre: configurator-scheduler-node-daemon
    spec:
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      initContainers:
      - name: fastpod-hook-init-ctn
        image: kontonpuku666/fastpod-hook-init:release
        volumeMounts:
        - name: fastpod-library
          mountPath: "/fastpod/library"
      containers:
      - name: fast-configurator-ctn
        image: kontonpuku666/fast-configurator:release
        volumeMounts:
        - name: fastpod-library
          mountPath: "/fastpod/library"
        - name: fastpod-scheduler
          mountPath: "/fastpod/scheduler"
        env:
        - name: FaSTPod-GPUClientsIP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
      - name: fastpod-fastscheduler-ctn
        image: kontonpuku666/fastpod-fastscheduler:release
        volumeMounts:
        - name: fastpod-library
          mountPath: "/fastpod/library"
        - name: fastpod-scheduler
          mountPath: "/fastpod/scheduler"
      volumes:
      - name: fastpod-library
        hostPath:
          path: "/fastpod/library"
      - name: fastpod-scheduler
        hostPath:
          path: "/fastpod/scheduler"
      
        

      

    


initContainers:
  - name: fast-hook-init-ctn
    image: kontonpuku666/fastpod-hook-init:release
    volumeMounts:
    - name: fastpod-library
      mountPath: "/fastpod/library"




cat > /etc/sysconfig/modules/ipvs.modules <<EOF
modprobe -- ip_vs
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- ip_vs_sh
modprobe -- nf_conntrack_ipv4
EOF


cat > /etc/sysconfig/modules/ipvs.modules <<EOF
 modprobe -- ip_vs 
 modprobe -- ip_vs_rr 
 modprobe -- ip_vs_wrr 
 modprobe -- ip_vs_sh 
 modprobe -- nf_conntrack
EOF


sudo chmod 755 /etc/sysconfig/modules/ipvs.modules && sudo bash /etc/sysconfig/modules/ipvs.modules && lsmod | sudo grep -e ip_vs -e nf_conntrack

kubectl get pod -n kube-system | grep kube-proxy |awk '{system("kubectl delete pod "$1" -n kube-system")}'

kubectl edit cm kube-proxy -n kube-system

kubectl get pod -n kube-system |grep kube-proxy
kubectl logs -n kube-system 







apiVersion: v1
kind: ServiceAccount
metadata:
  name: fastpod-controller-manager
  namespace: kube-system

---

kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fastpod-controller-manager-role
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["fastgshare.caps.in.tum"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["fastgshare.caps.in.tum"]
  resources: ["fastpods"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods", "pods/log", "namespaces", "endpoints"]
  verbs: ["get", "list", "watch"]



---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: fastpod-controller-manager-role-binder
subjects:
- kind: ServiceAccount
  name: fastpod-controller-manager
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: fastpod-controller-manager-role
  apiGroup: rbac.authorization.k8s.io

---

apiVersion: v1
kind: Service
metadata:
  name: fastpod-controller-manager-svc
  namespace: kube-system
spec:
  type: ClusterIP
  selector:
    app: fastpod-controller-manager
  ports:
  - protocol: TCP
    port: 10086
    targetPort: 10086

---

apiVersion: v1
kind: Pod
metadata:
  name: fastpod-controller-manager
  namespace: kube-system
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/path: "/metrics"
    prometheus.io/port: "9090"
  labels:
    app: fastpod-controller-manager
spec:
  serviceAccountName: fastpod-controller-manager
  priorityClassName: system-node-critical
  tolerations:
  - key: "CriticalAddonsOnly"
    operator: "Exists"
  - key: "node-role.kubernetes.io/master"
    operator: "Exists"
    effect: "NoSchedule"
  - key: "node-role.kubernetes.io/control-plane"
    operator: "Exists"
    effect: "NoSchedule"
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "node-role.kubernetes.io/master"
            operator: "Exists"
  restartPolicy: Always
  containers:
  - name: fastpod-controller-manager-container
    image: kontonpuku666/fastpod-controller-manager:release
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: kube-config
      mountPath: /root/.kube/config
      subPath: config
  volumes:
    - name: kube-config
      configMap:
        name: kube-config





