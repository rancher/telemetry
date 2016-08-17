FROM golang:1.6
ADD . /go/src/github.com/rancher/telemetry
RUN go get github.com/imikushin/trash
RUN trash
RUN go build github.com/rancher/telemetry
ENTRYPOINT ["/go/telemetry"]
CMD ["server"]
EXPOSE 8115
