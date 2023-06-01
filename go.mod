module github.com/rancher/telemetry

go 1.19

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/lib/pq v1.2.0
	github.com/rancher/norman v0.0.0-20210709145327-afd06f533ca3
	github.com/rancher/rancher/pkg/client v0.0.0-20210803030430-30574c2a9978
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	github.com/urfave/cli v1.20.0
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rancher/wrangler v1.1.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.25.4 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
)
