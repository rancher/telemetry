module github.com/rancher/telemetry

go 1.14

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/lib/pq v1.2.0
	github.com/rancher/norman v0.0.0-20210709145327-afd06f533ca3
	github.com/rancher/rancher/pkg/client v0.0.0-20210803030430-30574c2a9978
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/cli v1.20.0
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
)
