package cmd

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	publish "github.com/rancher/telemetry/publish"
	record "github.com/rancher/telemetry/record"
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

			cli.StringFlag{
				Name:   "pg-host",
				Usage:  "postgres host",
				Value:  "localhost",
				EnvVar: "TELEMETRY_PG_HOST",
			},
			cli.StringFlag{
				Name:   "pg-port",
				Usage:  "postgres port",
				Value:  "5432",
				EnvVar: "TELEMETRY_PG_PORT",
			},
			cli.StringFlag{
				Name:   "pg-user",
				Usage:  "postgres user",
				Value:  "telemetry",
				EnvVar: "TELEMETRY_PG_USER",
			},
			cli.StringFlag{
				Name:   "pg-pass",
				Usage:  "postgres password",
				Value:  "",
				EnvVar: "TELEMETRY_PG_PASS",
			},
			cli.StringFlag{
				Name:   "pg-dbname",
				Usage:  "postgres dbname",
				Value:  "telemetry",
				EnvVar: "TELEMETRY_PG_DBNAME",
			},
			cli.StringFlag{
				Name:   "pg-ssl",
				Usage:  "postgres ssl mode (disable, require, verify-ca, verify-full)",
				Value:  "disable",
				EnvVar: "TELEMETRY_PG_SSL",
			},

			cli.StringFlag{
				Name:   "admin-key",
				Usage:  "admin access key",
				Value:  "",
				EnvVar: "TELEMETRY_API_KEY",
			},

			cli.StringFlag{
				Name:   "admin-secret",
				Usage:  "admin secret key",
				Value:  "",
				EnvVar: "TELEMETRY_SECRET_KEY",
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
	router.HandleFunc("/healthcheck.html", serverCheck).Methods("GET")
	router.HandleFunc("/publish", serverPublish).Methods("POST")

	user := c.String("admin-key")
	pass := c.String("admin-secret")
	if user == "" || pass == "" {
		log.Warn("admin-{key,-secret} not set, admin disabled")
	} else {
		admin := mux.NewRouter()
		admin.HandleFunc("/admin/latest", adminApiLatest)
		admin.HandleFunc("/admin/by-day", adminApiByDay)
		admin.HandleFunc("/admin/install/{uid}", adminApiByUid)
		admin.HandleFunc("/admin/record/{id}", adminApiByRecordId)
		authed := httpauth.SimpleBasicAuth(user, pass)(admin)

		router.Handle("/admin", authed)
		router.Handle("/admin/{_dummy:.*}", authed)
	}

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	listen := c.String("listen")
	log.Info("Listening on ", listen)
	log.Fatal(http.ListenAndServe(listen, loggedRouter))
	return nil
}

func serverCheck(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pageok"))
}

func serverPublish(w http.ResponseWriter, req *http.Request) {
	var r record.Record

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&r)
	if err != nil {
		respondError(w, req, "Error parsing Record", 400)
		return
	}

	realIp := requestIp(req)
	ip := anonymizeIp(realIp)
	log.Debugf("Publish from %s: %s", realIp, r)

	err = googlePublisher.Report(r, ip)
	if err != nil {
		log.Errorf("Error publishing to Google: %s", err)
	}

	dbPublisher.Report(r, ip)
	if err != nil {
		log.Errorf("Error publishing to DB: %s", err)
	}

	respondSuccess(w, req, map[string]string{"ok": "1"})
}

// ------------

func adminApiLatest(w http.ResponseWriter, req *http.Request) {
	hours := 7
	hourStr := req.URL.Query().Get("hours")
	if hourStr != "" {
		hourInt, err := strconv.Atoi(hourStr)
		if err == nil {
			hours = hourInt
		}
	}

	if hours < 1 {
		respondError(w, req, "Hours is required", 422)
		return
	}

	installs, err := dbPublisher.GetLatest(hours)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	coll := Collection{
		Type:         "collection",
		ResourceType: "installation",
		Data:         installs,
	}

	respondSuccess(w, req, coll)
}

func adminApiByDay(w http.ResponseWriter, req *http.Request) {
	hours := 28 * 24
	hourStr := req.URL.Query().Get("hours")
	if hourStr != "" {
		hourInt, err := strconv.Atoi(hourStr)
		if err == nil {
			hours = hourInt
		}
	}

	if hours < 1 {
		respondError(w, req, "Hours is required", 422)
		return
	}

	data, err := dbPublisher.GetByDay(hours)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func adminApiByUid(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uid := vars["uid"]

	hours := 28 * 24
	hourStr := req.URL.Query().Get("hours")
	if hourStr != "" {
		hourInt, err := strconv.Atoi(hourStr)
		if err == nil {
			hours = hourInt
		}
	}

	if uid == "" {
		respondError(w, req, "UID is required", 422)
		return
	}

	if hours < 1 {
		respondError(w, req, "Hours is required", 422)
		return
	}

	records, err := dbPublisher.GetRecordsByUid(uid, hours)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	coll := Collection{
		Type:         "collection",
		ResourceType: "record",
		Data:         records,
	}

	respondSuccess(w, req, coll)
}

func adminApiByRecordId(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	if id == "" {
		respondError(w, req, "ID is required", 422)
		return
	}

	record, err := dbPublisher.GetRecordById(id)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, record)
}

func adminUi(w http.ResponseWriter, req *http.Request) {
	respondSuccess(w, req, "<html><body>Hi</body></html>")
}

// ------------

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

func anonymizeIp(in string) string {
	ip := net.ParseIP(in).To16()
	if ip == nil {
		return in
	}

	var mask net.IPMask
	v4 := ip.To4()
	if v4 == nil {
		mask = net.CIDRMask(8*8, 8*16)
	} else {
		mask = net.CIDRMask(8*3, 8*4)
	}

	return ip.Mask(mask).String()
}
