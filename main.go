package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/urfave/cli"

	rancher "github.com/rancher/go-rancher/client"
	collector "github.com/vincent99/telemetry/collector"
	record "github.com/vincent99/telemetry/record"
)

const UID_FILE = ".telemetry_id"

var (
	router = mux.NewRouter()

	VERSION string
	ticker  = time.NewTicker(1 * time.Minute)

	url       string
	accessKey string
	secretKey string
)

func main() {
	app := cli.NewApp()
	app.Name = "telemetry"
	app.Usage = "Rancher telemetry daemon"
	app.Version = VERSION
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug logging",
			EnvVar: "TELEMETRY_DEBUG",
		},

		cli.StringFlag{
			Name:   "listen, l",
			Usage:  "address/port to listen on",
			Value:  "0.0.0.0:8114",
			EnvVar: "TELEMETRY_LISTEN",
		},

		cli.StringFlag{
			Name:   "log",
			Usage:  "path to log to",
			Value:  "",
			EnvVar: "TELEMETRY_LOG",
		},

		cli.StringFlag{
			Name:   "pid-file",
			Usage:  "path to write PID to",
			Value:  "",
			EnvVar: "TELEMETRY_PID_FILE",
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
	}

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	log.Infof("Telemetry %s", VERSION)

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	logFile := c.String("log")
	if logFile != "" {
		if output, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			str := fmt.Sprintf("Failed to log to file %s: %v", logFile, err)
			return cli.NewExitError(str, 1)
		} else {
			log.SetOutput(output)
		}
	}

	pidFile := c.String("pid-file")
	if pidFile != "" {
		log.Infof("Writing pid %d to %s", os.Getpid(), pidFile)
		if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
			str := fmt.Sprintf("Failed to write pid file %s: %v", pidFile, err)
			return cli.NewExitError(str, 1)
		}
	}

	if url == "" || accessKey == "" || secretKey == "" {
		return cli.NewExitError("URL, Access Key, and Secret Key are required", 1)
	}

	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/", show).Methods("GET")
	router.HandleFunc("/", reload).Methods("POST")

	listen := c.String("listen")
	log.Info("Listening on ", listen)
	log.Fatal(http.ListenAndServe(listen, router))
	return nil
}

func respondError(w http.ResponseWriter, req *http.Request, msg string, statusCode int) {
	obj := make(map[string]interface{})
	obj["message"] = msg
	obj["type"] = "error"
	obj["code"] = statusCode

	bytes, err := json.Marshal(obj)
	if err == nil {
		http.Error(w, string(bytes), statusCode)
	} else {
		http.Error(w, "{\"type\": \"error\", \"message\": \"JSON marshal error\"}", http.StatusInternalServerError)
	}
}

func respondSuccess(w http.ResponseWriter, req *http.Request, val interface{}) {
	bytes, err := json.Marshal(val)
	if err == nil {
		w.Write(bytes)
	} else {
		respondError(w, req, "Error serializing to JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func reload(w http.ResponseWriter, req *http.Request) {
}

func show(w http.ResponseWriter, req *http.Request) {
	r, err := collect()
	if err == nil {
		respondSuccess(w, req, r)
	} else {
		respondError(w, req, err.Error(), 500)
	}
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
	r["v"] = 1                                    // Record version
	r["uid"] = getUid()                           // Installation UID
	r["ts"] = time.Now().UTC().Round(time.Second) // Local time

	opt := collector.CollectorOpts{
		Client: client,
	}

	collector.Run(&r, &opt)

	return r, nil
}

func getUid() string {
	var id string

	data, err := ioutil.ReadFile(UID_FILE)
	if err == nil {
		id = string(data)
	} else {
		id = uuid.NewV4().String()
		ioutil.WriteFile(UID_FILE, []byte(id), 0644)
	}

	return id
}
