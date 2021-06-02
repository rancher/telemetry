FROM alpine:3.11.11
RUN apk -u --no-cache add curl ca-certificates
ADD "bin/telemetry" /
ENTRYPOINT ["/telemetry"]
CMD ["server"]
EXPOSE 8115
