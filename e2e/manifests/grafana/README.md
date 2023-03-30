# Grafana

This folder uses a `kustomization.yaml` file to generate a config map from the
files in the `telemetry` subfolder.

The resulting config map is too large to be applied client-side, where a
`metadata.annotations` field is added that contains the previous content of the
config map. It simply exceeds the 262144 max bytes it may have.

Instead of the usual apply operating, use the server-side apply mechanism:
`kubectl apply --server-side`, which moves the logic of determining changes in
the manifest to the server and hence also doesn't create a
`last-applied-configuration` annotation. You still need to keep using the `-k`
flag to tell `kubectl` to run `kustomize`.

Example:

```console
kubectl apply --server-side -k manifests/grafana
```
