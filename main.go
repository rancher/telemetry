package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/vincent99/telemetry/collector"
)

var (
	showVersion = flag.Bool("version", false, "Show version")
	debug       = flag.Bool("debug", false, "Debug")
	listen      = flag.String("listen", "127.0.0.1:8114", "Address to listen to (TCP)")
	logFile     = flag.String("log", "", "Log file")
	pidFile     = flag.String("pid-file", "", "PID to write to")

	router = mux.NewRouter()

	VERSION string
	ticker  = time.NewTicker(1 * time.Minute)
)

func main() {
	parseFlags()

	if *showVersion {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	log.Infof("Starting telemetry %s", VERSION)

	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/", show).Methods("GET")
	router.HandleFunc("/", reload).Methods("POST")

	log.Info("Listening on ", *listen)
	log.Fatal(http.ListenAndServe(*listen, router))
}

func parseFlags() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *logFile != "" {
		if output, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			log.Fatalf("Failed to log to file %s: %v", *logFile, err)
		} else {
			log.SetOutput(output)
		}
	}

	if *pidFile != "" {
		log.Infof("Writing pid %d to %s", os.Getpid(), *pidFile)
		if err := ioutil.WriteFile(*pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
			log.Fatalf("Failed to write pid file %s: %v", *pidFile, err)
		}
	}
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

func collect() (*Record, error) {
	r := &Record{
		Version:      1,
		Installation: collector.GetInstallation(),
		Os:           collector.GetOs(),
	}

	return r, nil
}
