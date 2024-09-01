FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG GIT_COMMIT

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

WORKDIR /go/src/github.com/KontonGu/FaST-GShare
COPY . .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")

RUN echo $CGO_ENABLED ${TARGETOS} ${TARGETARCH}

RUN  go build -o fast-gshare-faas .



FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.19 as ship
LABEL org.label-schema.license="Apache 2.0" \
      org.label-schema.vcs-url="https://github.com/KontonGu/FaST-GShare" \
      org.label-schema.vcs-type="Git" \
      org.label-schema.name="fastgshare.caps.in.tum/fast-gshare" \
      org.label-schema.vendor="fast-gshare" \
      org.label-schema.docker.schema-version="1.0" \
      org.opencontainers.image.source="https://github.com/KontonGu/FaST-GShare"

RUN apk --no-cache add ca-certificates

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/KontonGu/FaST-GShare/fast-gshare-faas .
RUN chown -R app:app ./

USER app

CMD ["./fast-gshare-faas"]
