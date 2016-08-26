package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	rancher "github.com/rancher/go-rancher/client"
	collector "github.com/rancher/telemetry/collector"
	publish "github.com/rancher/telemetry/publish"
	record "github.com/rancher/telemetry/record"
)

var (
	publisher *publish.ToUrl
	url       string
	accessKey string
	secretKey string
	target    string
)

const RECORD_VERSION = 1
const EXISTING_FILE = ".existing"

func ClientCommand() cli.Command {
	return cli.Command{
		Name:   "client",
		Usage:  "report stats to a telemetry server",
		Action: clientRun,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "once",
				Usage: "print stats to stdout once and exit",
			},

			cli.StringFlag{
				Name:   "listen, l",
				Usage:  "address/port to listen on",
				Value:  "0.0.0.0:8114",
				EnvVar: "TELEMETRY_LISTEN",
			},

			cli.StringFlag{
				Name:        "url",
				Usage:       "url to reach cattle",
				Value:       "",
				EnvVar:      "CATTLE_URL",
				Destination: &url,
			},

			cli.StringFlag{
				Name:        "access-key",
				Usage:       "access key for api",
				Value:       "",
				EnvVar:      "CATTLE_ACCESS_KEY",
				Destination: &accessKey,
			},

			cli.StringFlag{
				Name:        "secret-key",
				Usage:       "secret key for api",
				Value:       "",
				EnvVar:      "CATTLE_SECRET_KEY",
				Destination: &secretKey,
			},

			cli.StringFlag{
				Name:   "interval",
				Usage:  "reporting interval",
				Value:  "6h",
				EnvVar: "TELEMETRY_INTERVAL",
			},

			cli.StringFlag{
				Name:   "to-url",
				Usage:  "url to send stats to",
				Value:  "https://telemetry.rancher.io/publish",
				EnvVar: "TELEMETRY_TO_URL",
			},
		},
	}
}

func clientRun(c *cli.Context) error {
	if c.Bool("once") {
		return clientShowOnce()
	}

	log.Infof("Telemetry Client %s", c.App.Version)

	if url == "" || accessKey == "" || secretKey == "" {
		return cli.NewExitError("URL, Access Key, and Secret Key are required", 1)
	}

	clientVersion := c.String("version")
	if clientVersion == "" {
		clientVersion = "unknown"
	}

	publisher = publish.NewToUrl(c)

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/v1-telemetry", clientShow).Methods("GET")
	router.HandleFunc("/v1-telemetry/reload", clientReload).Methods("POST")
	router.HandleFunc("/v1-telemetry/report", clientReport).Methods("POST")

	interval := c.String("interval")
	if interval != "" {
		dur, err := time.ParseDuration(interval)
		if err != nil {
			return cli.NewExitError("Interval must be a valid GoLang duration string", 1)
		}

		if dur.Nanoseconds() > 0 {
			ticker := time.NewTicker(dur)
			go func() {
				for {
					select {
					case <-ticker.C:
						report()
					}
				}
			}()
		}
	}

	// Report immediately on only the first run
	if !isExisting() {
		go report()
	}

	listen := c.String("listen")
	log.Info("Listening on ", listen)
	log.Fatal(http.ListenAndServe(listen, router))
	return nil
}

// CLI Handlers
func clientShowOnce() error {
	r, err := collect()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	str, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Print(string(str))
	fmt.Print("\n")
	return cli.NewExitError("", 0)
}

// HTTP Handlers
func clientShow(w http.ResponseWriter, req *http.Request) {
	r, err := collect()
	if err == nil {
		respondSuccess(w, req, r)
	} else {
		respondError(w, req, err.Error(), 500)
	}
}

func clientReload(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok"))
}

func clientReport(w http.ResponseWriter, req *http.Request) {
	report()
	w.Write([]byte("ok"))
}

func report() {
	start := time.Now()
	log.Debug("Starting report")

	r, err := collect()
	if err != nil {
		log.Errorf("Error collecting data: %s", err)
		return
	}
	diff := time.Now().Sub(start).String()
	log.Debugf("Collected stats in %s", diff)

	err = publisher.Report(r, "")
	if err != nil {
		log.Errorf("Error publishing report: %s", err)
		return
	}

	diff = time.Now().Sub(start).String()
	log.Debugf("Completed report in %s", diff)
}

func collect() (record.Record, error) {
	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})

	if err != nil {
		return nil, err
	}

	r := record.Record{}
	r["r"] = RECORD_VERSION
	r["ts"] = time.Now().UTC().Format(time.RFC3339)

	opt := collector.CollectorOpts{
		Client: client,
	}

	collector.Run(&r, &opt)

	return r, nil
}

func isExisting() bool {
	want := strconv.Itoa(RECORD_VERSION)
	have := ""

	data, err := ioutil.ReadFile(EXISTING_FILE)
	if err == nil {
		have = string(data)
	}

	if want == have {
		return true
	} else {
		ioutil.WriteFile(EXISTING_FILE, []byte(want), 0644)
		return false
	}
}
