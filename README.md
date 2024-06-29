# FaST-GShare: Enabling Efficient Spatio-Temporal GPU Sharing in Serverless Computing for Deep Learning Inference

## Introduction

## Deploy



## Build
1. Get Go dependencies
    ```
    $ go mod tidy
    ```
2. Generate FaSTPod Custom Resource Definition (CRD) Client
    ```
    $ go get k8s.io/code-generator@v0.24.2
    $ bash code-gen.sh
    ```

