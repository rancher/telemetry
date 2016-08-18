FROM ubuntu:16.04
ADD "bin/telemetry" /
ENTRYPOINT ["/telemetry"]
CMD ["server"]
EXPOSE 8115
