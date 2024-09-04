# FaST-GShare: Enabling Efficient Spatio-Temporal GPU Sharing in Serverless Computing for Deep Learning Inference

## Introduction

### FaSTPod 
FaSTPod is a Custom Resource that enables temporal and spacial GPU sharing. Users can specify fine-grained GPU resources for each Pod in the FaSTPod's YAML specification annotations, such as:
```
annotations: 
  fastgshare/gpu_quota_request: "0.7"
  fastgshare/gpu_quota_limit: "0.8"
  fastgshare/gpu_sm_partition: "30"
  fastgshare/gpu_mem: "2700000000"
```
Additionally, users can define the required replicas:
```
spec:
  replicas: 2
```
A simple FaSTPod deployment example is available in `yaml/fastpod/testfastpod.yaml`.



## Deployment

### Infrastructure Install

### Deploy FaST-GShare FaSTPod 
1. Deploy FaSTPod CRD (Custom Resource Definition)
    ```
    $ kubectl apply -f ./yaml/crds/fastpod_crd.yaml
    ```
2. Deploy FaSTPod Controller Manager and GPU Resource Configurator
    ```
    $ bash ./yaml/fastgshare/apply_deploy_ctr_mgr_node_daemon.sh
    ```
3. Test the FaSTPod example    
    ```
    $ kubectl apply -f ./yaml/fastpod/testfastpod.yaml
    ```
    check the FaSTPods and correponding Pods deployed:
    ```
    $ kubectl get pods -n fast-gshare
    $ kubectl get fastpods -n fast-gshare
    ```
---
### Deploy FaST-GShare-Function (without Autoscaler)
The deployment of the FaST-GShare-Function in this project does not include the FaST-GShare-Autoscaler and only deploys with a fixed number of replicas. The complete FaST-GShare serverless version is available at [FaST-GShare-Function](https://github.com/KontonGu/FaST-GShare-Function), and the basic verion of Autoscaler plugin can be found at [FaST-GShare-Autoscaler](https://github.com/KontonGu/FaST-GShare-Autoscaler.git).
1. Install the FaST-GShare-Function Components
    ```
    $ bash ./install/fast-gshare-fn-deploy.sh
    ```

2. Uninstall the FaST-GShare-Function Components
    ```
    $ make helm_uninstall_fast-gshare-fn
    ```

## Build FaST-GShare from Scrach based on the Code (for further Developing)
1. Get Go dependencies
    ```
    $ mkdir github.com/KontonGu && cd github.com/KontonGu
    $ git clone git@github.com:KontonGu/FaST-GShare.git
    $ cd FaST-GShare
    [optional]
    $ go mod init github.com/KontonGu/FaST-GShare
    $ go mod tidy
    ```
2. Generate FaSTPod Custom Resource Definition (CRD) Client
    Clean original generated FaSTPod CRD client
    ```
    $ make clean_crd_gen
    ```
    Generate new CRD client:
    ```
    $ make code_gen_crd
    ```
    Update CRD client:
    ```
    $ make update_crd
    ```



## Citation
If you use FaST-GShare for your research, please cite our, please cite our paper [paper](https://dl.acm.org/doi/abs/10.1145/3605573.3605638):
```
@inproceedings{gu2023fast,
  title={FaST-GShare: Enabling Efficient Spatio-Temporal GPU Sharing in Serverless Computing for Deep Learning Inference},
  author={Gu, Jianfeng and Zhu, Yichao and Wang, Puxuan and Chadha, Mohak and Gerndt, Michael},
  booktitle={Proceedings of the 52nd International Conference on Parallel Processing},
  pages={635--644},
  year={2023}
}
```
