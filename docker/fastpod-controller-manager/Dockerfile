FROM ubuntu:22.04 AS build

ENV GOLANG_VERSION 1.22.0
ENV PATH $PATH:/usr/local/go/bin

WORKDIR /go/src/github.com/KontonGu/FaST-GShare
COPY . .

RUN apt-get update && apt install -y g++ wget make && \
    wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    make fastpod-controller-manager


# FROM busybox:1.36.1-glibc
FROM  ubuntu:22.04

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=utility

COPY --from=build /go/src/github.com/KontonGu/FaST-GShare/cmd/fastpod-controller-manager/fastpodcontrollermanager /usr/bin/fastpodcontrollermanager

# CMD ["fastpodcontrollermanager", "-alsologtostderr", "-v=4"]
CMD [ "fastpodcontrollermanager", "-alsologtostderr"]









