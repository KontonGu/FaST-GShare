FROM nvidia/cuda:12.2.0-devel-ubuntu22.04 AS build

WORKDIR /tmp/build

COPY docker/dummyPod .

RUN apt-get update && \
    apt install -y build-essential && \
    make build


FROM busybox:1.36.1-glibc

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=utility


COPY --from=build /usr/lib/x86_64-linux-gnu/libdl.so.2 /lib
COPY --from=build /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /lib
COPY --from=build /usr/lib/x86_64-linux-gnu/librt.so.1 /lib
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcc_s.so.1 /lib
COPY docker/libs/libnvidia-ml.so.1 /lib


COPY --from=build /tmp/build/get_uuid /usr/bin/get_uuid
CMD ["/usr/bin/get_uuid"]



