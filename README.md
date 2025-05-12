## Deployment

### Infrastruction Install
Install K8S infrastructure, NVIDIA Driver && Toolkit, and other prerequisite, please follow [Installation Guide](https://github.com/KontonGu/k8s-cluster-config).

### Install FaSTPod 
1. Deploy FaSTPod CRD (Custom Resource Definition) [Required]
    ```
    $ kubectl apply -f ./yaml/crds/fastpod_crd.yaml
    ```
2. Deploy FaSTPod Controller Manager and GPU Resource Configurator [Required]
    ```
    $ bash ./yaml/fastgshare/apply_deploy_ctr_mgr_node_daemon.sh
    ```
3. Test the FaSTPod example [Optional, to validate the proper functioning of FastPod]   
    ```
    $ kubectl apply -f ./yaml/fastpod/testfastpod.yaml
    ```
    check the FaSTPods and correponding Pods deployed:
    ```
    $ kubectl get pods -n fast-gshare
    $ kubectl get fastpods -n fast-gshare
    ```
4. Uninstall the FaST-GShare FaSTPod deployment [Optional, only when we need to uninstall]
    ```
    $ bash ./yaml/fastgshare/clean_deploy_ctr_mgr_node_daemon.sh
    ```
5. If the kube-config is missed [Required] [Optional when configmap exists]
    ```
    $ kubectl create configmap kube-config -n kube-system --from-file=$HOME/.kube/config
    ```
   
---
### Install and Uninstall FaST-GShare-Function (without Autoscaler)
The deployment of the FaST-GShare-Function in this project does not include the FaST-GShare-Autoscaler and only deploys with a fixed number of replicas. The complete FaST-GShare serverless version should include the FaST-GShare-Autoscaler, and the basic verion of Autoscaler plugin can be found at [FaST-GShare-Autoscaler](https://github.com/KontonGu/FaST-GShare-Autoscaler.git).
- Install the FaST-GShare-Function Components
    ```
    $ bash ./install/install_fast-gshare-fn.sh
    ```
    Test if the FaST-GShare-Function is successfully deployed:
    ```
    $ kubectl apply -f yaml/fastpod/test-fastgshare-fn.yaml
    ```
- Uninstall the FaST-GShare-Function Components
    ```
    $ bash ./install/uninstall_fast-gshare-fn.sh
    ```

## Build FaST-GShare from Scratch based on the Code (for further Developing)
The detailed introduction to the FaST-GShare project's construction from the source code can be found in the `./develope` directory and [README](https://github.com/KontonGu/FaST-GShare/blob/main/develop/README.md).



## License
Copyright 2024 FaST-GShare Authors, KontonGu (**Jianfeng Gu**), et. al.
@Techinical University of Munich, **CAPS Cloud Team**

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
