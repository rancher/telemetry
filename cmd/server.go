package cmd

import (
	"encoding/json"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	publish "github.com/vincent99/telemetry/publish"
	record "github.com/vincent99/telemetry/record"
)

var (
	enableXff       bool
	googlePublisher *publish.Google
	dbPublisher     *publish.Postgres
)

func ServerCommand() cli.Command {
	return cli.Command{
		Name:   "server",
		Usage:  "gather stats from a telemetry client",
		Action: serverRun,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "listen, l",
				Usage: "address/port to listen on",
				Value: "0.0.0.0:8115",
			},

			cli.BoolFlag{
				Name:        "xff",
				Usage:       "enable support for X-Forwarded-For header",
				Destination: &enableXff,
			},

			cli.StringFlag{
				Name:   "ga-tid",
				Usage:  "google analytics tracking id",
				Value:  "",
				EnvVar: "TELEMETRY_GA_TID",
			},
		},
	}
}

func serverRun(c *cli.Context) error {
	log.Infof("Telemetry Server %s", c.App.Version)

	googlePublisher = publish.NewGoogle(c)
	dbPublisher = publish.NewPostgres(c)

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/reload", serverReload).Methods("POST")
	router.HandleFunc("/publish", serverPublish).Methods("POST")

	listen := c.String("listen")
	log.Info("Listening on ", listen)
	log.Fatal(http.ListenAndServe(listen, router))
	return nil
}

func serverReload(w http.ResponseWriter, req *http.Request) {
	respondSuccess(w, req, "OK")
}

func serverPublish(w http.ResponseWriter, req *http.Request) {
	var r record.Record

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&r)
	if err != nil {
		respondError(w, req, "Error parsing Record", 400)
		return
	}

	ip := requestIp(req)
	log.Debugf("Publish from %s: %s", ip, r)

	err = googlePublisher.Report(r, ip)
	if err != nil {
		log.Errorf("Error publishing to Google: %s", err)
	}

	dbPublisher.Report(r, ip)
	if err != nil {
		log.Errorf("Error publishing to DB: %s", err)
	}

	respondSuccess(w, req, "OK")
}

func requestIp(req *http.Request) string {
	if enableXff {
		clientIp := req.Header.Get("X-Forwarded-For")
		if len(clientIp) > 0 {
			return clientIp
		}
	}

	clientIp, _, _ := net.SplitHostPort(req.RemoteAddr)
	return clientIp
}
