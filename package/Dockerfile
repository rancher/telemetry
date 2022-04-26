FROM alpine:3.14.3
RUN apk -u --no-cache add curl ca-certificates
ADD "bin/telemetry" /
ENTRYPOINT ["/telemetry"]
CMD ["server"]
EXPOSE 8115
