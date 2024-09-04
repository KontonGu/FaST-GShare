## Build FaST-GShare from Scrach based on the Code (for further Developing)

### FaSTPod
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

### fastpod-controller-manager
rebuild && upload the fastpod-controller-manager docker image:
```
$ make test-build-fastpod-controller-manager-container 
```


### fast-configurator
rebuild && upload the fast-configurator docker image:
```
$ make test-fast-configurator-container
```

### fast-gshare-fn
rebuild && upload the fast-gshare-fn container image:
```
$ make test-fast-gshare-faas-imag
```