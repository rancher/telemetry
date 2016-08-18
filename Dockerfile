FROM ubuntu:16.04
RUN apt-get update && apt-get dist-upgrade -y && apt-get install -y ca-certificates
ADD "bin/telemetry" /
ENTRYPOINT ["/telemetry"]
CMD ["server"]
EXPOSE 8115
