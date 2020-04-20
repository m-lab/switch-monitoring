FROM golang:1.13 as build
ENV CGO_ENABLED 0
ADD . /go/src/github.com/m-lab/switch-monitoring/cmd/switch-monitoring
WORKDIR /go/src/github.com/m-lab/switch-monitoring/cmd/switch-monitoring
RUN go get \
    -v \
    github.com/m-lab/switch-monitoring/cmd/switch-monitoring

# Now copy the built image into the minimal base image
FROM alpine:3.10
COPY --from=build /go/bin/switch-monitoring /
WORKDIR /
ENTRYPOINT ["/switch-monitoring"]
