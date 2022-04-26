module github.com/rancher/telemetry

go 1.16

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/rancher/norman => github.com/rancher/norman v0.0.0-20211201154850-abe17976423e
	k8s.io/client-go => github.com/rancher/client-go v1.20.0-rancher.1
	github.com/rancher/wrangler => github.com/rancher/wrangler v0.8.11-0.20211214201934-f5aa5d9f2e81
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20220426064615-7db4fe4e8193
)

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/lib/pq v1.2.0
	github.com/rancher/norman v0.0.0-20211201154850-abe17976423e
	github.com/rancher/rancher/pkg/client v0.0.0-20220426064615-7db4fe4e8193
	github.com/rancher/wrangler v0.8.11-0.20211214201934-f5aa5d9f2e81
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/cli v1.22.2
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
	k8s.io/api v0.20.0 // indirect
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v12.0.0+incompatible // indirect
	k8s.io/kube-openapi v0.0.0-20211110013926-83f114cd0513 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
)
