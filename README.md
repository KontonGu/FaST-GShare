# FaST-GShare: Enabling Efficient Spatio-Temporal GPU Sharing in Serverless Computing for Deep Learning Inference

## Introduction

## Deploy



## Build
1. Get Go dependencies
    ```
    $ mkdir github.com/KontonGu && cd github.com/KontonGu
    $ git clone git@github.com:KontonGu/FaST-GShare.git
    $ cd FaST-GShare
    $ go mod init github.com/KontonGu/FaST-GShare
    $ go mod tidy
    ```
2. Generate FaSTPod Custom Resource Definition (CRD) Client
    ```
    $ cd ../
    $ git clone https://github.com/kubernetes/code-generator.git
    $ git checkout release-1.23 && cd ../FaST-GShare
    $ bash code-gen.sh
    ```

