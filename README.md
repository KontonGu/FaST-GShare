# FaST-GShare: Enabling Efficient Spatio-Temporal GPU Sharing in Serverless Computing for Deep Learning Inference

## Introduction

## Deploy



## Build
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
    $ make gen_crd
    ```

