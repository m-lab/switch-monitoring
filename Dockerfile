FROM golang:1.20 as build
ENV CGO_ENABLED 0
ADD . /go/src/github.com/m-lab/switch-monitoring
WORKDIR /go/src/github.com/m-lab/switch-monitoring
RUN go install -v ./cmd/switch-monitoring

# Now copy the built image into the minimal base image
FROM alpine:3.10
COPY --from=build /go/bin/switch-monitoring /
WORKDIR /
ENTRYPOINT ["/switch-monitoring"]
