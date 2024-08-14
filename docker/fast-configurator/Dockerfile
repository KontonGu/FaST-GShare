FROM nvidia/cuda:12.2.2-base-ubuntu22.04 AS build

ENV GOLANG_VERSION 1.22.0
ENV PATH $PATH:/usr/local/go/bin

WORKDIR /go/src/github.com/KontonGu/FaST-GShare
COPY . .

RUN apt-get update && apt install -y g++ wget make && \
    wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    make fast-configurator

FROM nvidia/cuda:12.2.2-base-ubuntu22.04

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILTIES=utility

COPY --from=build /go/src/github.com/KontonGu/FaST-GShare/cmd/fast-configurator/fast-configurator /usr/bin/fast-configurator

CMD [ "fast-configurator", "-alsologtostderr", "-ctr_mgr_ip_port=fastpod-controller-manager-svc.kube-system.svc.cluster.local:10086", "-v=4"]
