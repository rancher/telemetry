package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/rancher/norman/clientbase"
	collector "github.com/rancher/telemetry/collector"
	publish "github.com/rancher/telemetry/publish"
	record "github.com/rancher/telemetry/record"
	rancher "github.com/rancher/types/client/management/v3"
)

const EXISTING_FILE = ".existing"

var (
	publisher  *publish.ToUrl
	licenser   *publish.ToUrl
	records    map[string]record.Record
	url        string
	accessKey  string
	secretKey  string
	tokenKey   string
	caCert     string
	target     string
	collecting sync.Mutex
)

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
				Name:        "token-key",
				Usage:       "token key for api",
				Value:       "",
				EnvVar:      "CATTLE_TOKEN_KEY",
				Destination: &tokenKey,
			},

			cli.StringFlag{
				Name:   "crt-file",
				Usage:  "Ca trust certificate file for api",
				Value:  "",
				EnvVar: "CATTLE_CERTIFICATE",
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
				Value:  "https://telemetry.rancher.io",
				EnvVar: "TELEMETRY_TO_URL",
			},
		},
	}
}

func clientRun(c *cli.Context) error {
	log.Infof("Telemetry Client %s", c.App.Version)

	if url == "" || (tokenKey == "" && (accessKey == "" || secretKey == "")) {
		return cli.NewExitError("URL, Access Key and Secret Key OR Token Key are required", 1)
	}

	url = normalizeURL(url)

	if tokenKey == "" {
		tokenKey = accessKey + ":" + secretKey
	}

	clientVersion := c.String("version")
	if clientVersion == "" {
		clientVersion = "unknown"
	}

	crt_file := c.String("crt-file")
	if crt_file != "" {
		crt, err := ioutil.ReadFile(crt_file)
		if err != nil {
			return cli.NewExitError("Error reading certificate file", 1)
		}
		caCert = string(crt)
	}

	if c.Bool("once") {
		return clientShowOnce()
	}

	publisher = publish.NewToUrl(c, PUBLISH_URI)
	licenser = publish.NewToUrl(c, LICENSE_URI)

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/v1-telemetry", clientShow).Methods("GET")
	router.HandleFunc("/v1-telemetry/reload", clientReload).Methods("POST")
	router.HandleFunc("/v1-telemetry/report", clientReport).Methods("POST")
	router.HandleFunc("/v1-license", licenseShow).Methods("GET")
	router.HandleFunc("/v1-license/check", licenseCheck).Methods("POST")

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
	err := collect()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	collecting.Lock()
	defer collecting.Unlock()

	for i := range records {
		str, err := json.MarshalIndent(records[i], "", "  ")
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		fmt.Print(string(str))
		fmt.Print("\n")
	}
	return cli.NewExitError("", 0)
}

// HTTP Handlers
func clientShow(w http.ResponseWriter, req *http.Request) {
	err := collect()
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	collecting.Lock()
	defer collecting.Unlock()

	if records[collector.RECORD_INSTALLATION] == nil {
		respondSuccess(w, req, "Telemetry is disabled")
		return
	}
	respondSuccess(w, req, records[collector.RECORD_INSTALLATION])
}

func licenseShow(w http.ResponseWriter, req *http.Request) {
	err := collect()
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	collecting.Lock()
	defer collecting.Unlock()

	if records[collector.RECORD_LICENSING] == nil {
		respondSuccess(w, req, "Rancher is not licensed")
		return
	}
	respondSuccess(w, req, records[collector.RECORD_LICENSING])
}

func licenseCheck(w http.ResponseWriter, req *http.Request) {
	err := collect()
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	collecting.Lock()
	defer collecting.Unlock()

	if records[collector.RECORD_LICENSING] == nil {
		respondSuccess(w, req, "No License data")
		return
	}

	resp, err := licenser.Report(records[collector.RECORD_LICENSING], "")
	if err != nil {
		log.Errorf("Error licensing report: %s", err)
		respondError(w, req, string(resp), 400)
		return
	}
	w.Write(resp)
}

func clientReload(w http.ResponseWriter, req *http.Request) {
	err := collect()
	if err != nil {
		log.Errorf("Error collecting data: %s", err)
		respondError(w, req, "Error reloading data", 400)
		return
	}
	w.Write([]byte("ok"))
}

func clientReport(w http.ResponseWriter, req *http.Request) {
	report()
	w.Write([]byte("ok"))
}

func report() {
	start := time.Now()
	log.Debug("Starting report")

	err := collect()
	if err != nil {
		log.Errorf("Error collecting data: %s", err)
		return
	}
	diff := time.Now().Sub(start).String()
	log.Debugf("Collected stats in %s", diff)

	collecting.Lock()
	defer collecting.Unlock()

	if records[collector.RECORD_INSTALLATION] != nil {
		_, err = publisher.Report(records[collector.RECORD_INSTALLATION], "")
		if err != nil {
			log.Errorf("Error publishing report: %s", err)
			return
		}
		diff = time.Now().Sub(start).String()
		log.Debugf("Published telemetry in %s", diff)
	}

	if records[collector.RECORD_LICENSING] != nil {
		_, err = licenser.Report(records[collector.RECORD_LICENSING], "")
		if err != nil {
			log.Errorf("Error licensing report: %s", err)
			return
		}
		diff = time.Now().Sub(start).String()
		log.Debugf("Published licensing in %s", diff)
	}

	diff = time.Now().Sub(start).String()
	log.Debugf("Completed report in %s", diff)
}

func collect() error {
	client, err := rancher.NewClient(&clientbase.ClientOpts{
		URL:      url,
		TokenKey: tokenKey,
		Insecure: true,
	})
	if err != nil {
		return err
	}

	opt := collector.CollectorOpts{
		Client: client,
	}

	collecting.Lock()
	defer collecting.Unlock()

	records = collector.Run(&opt)

	return nil
}

func isExisting() bool {
	want := strconv.Itoa(collector.RECORD_VERSION)
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

func normalizeURL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSuffix(url, "/")

	if !strings.HasSuffix(url, "/v3") {
		url = url + "/v3"
	}

	return url
}
