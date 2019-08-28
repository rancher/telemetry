# Telemetry

Telemetry server and client for Rancher


## Building

`make`

## Developing the client

Have a local rancher server running, [see wiki for setup details](https://github.com/rancher/rancher/wiki).

In your rancher instance create [an api key](https://localhost:8443/apikeys) and copy the bearer token. 

Now fork the telemetry repo and clone in a separate directory:

```
git clone https://github.com/rancher/telemetry.git  # use your fork
cd telemetry
go run main.go client --url=https://localhost:8443/v3 --token-key=token-abc:xyz
```

Now there are a few ways to hit your client: 

* Hit the client server directly: http://localhost:8114/v1-telemetry
* Hit via rancher at https://localhost:8443/v1-telemetry
* Instead of running a server you can use the 'once' param: `--once | jq '.cluster.pod'`

