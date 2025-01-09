## Infrastructure Install

### Install NVIDIA Driver
```
$ bash install_nvidia_driver.sh
```

### Install Kubernetes Nodes
- Install the master node:
    ```
    $ bash install_k8s_master_node.sh
    $ bash install_nvidia_container_toolkit.sh
    ```
- Install work nodes:
    ```
    $ bash install_k8s_work_node.sh
    $ bash install_nvidia_container_toolkit.sh
    ```




### Other Prerequisite
You can use the following bash script to install all prerequisites with one click.
```
$ bash prerequisite_install.sh
```
- `helm` for FaST-GShare Deployment
- `k6` for load testing
