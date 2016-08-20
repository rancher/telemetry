package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	publish "github.com/rancher/telemetry/publish"
	record "github.com/rancher/telemetry/record"
)

var (
	version         string
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
	rand.Seed(time.Now().UnixNano())

	version = c.App.Version
	googlePublisher = publish.NewGoogle(c)
	dbPublisher = publish.NewPostgres(c)

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", http.NotFound)
	router.HandleFunc("/healthcheck.html", serverCheck).Methods("GET")
	router.HandleFunc("/publish", serverPublish).Methods("POST")
	router.HandleFunc("/", serverRoot).Methods("GET")

	user := c.String("admin-key")
	pass := c.String("admin-secret")
	if user == "" || pass == "" {
		log.Warn("admin-{key,-secret} not set, admin disabled")
	} else {
		admin := mux.NewRouter()
		admin.HandleFunc("/admin/installs", apiActiveInstalls)                   // ?hours=7
		admin.HandleFunc("/admin/counts/{fields}", apiLatestCounts)              // ?hours=7
		admin.HandleFunc("/admin/count-map/{field}", apiLatestCountMap)          // ?hours=7
		admin.HandleFunc("/admin/historical/{fields}", apiHistoricalCounts)      // ?days=28
		admin.HandleFunc("/admin/historical-map/{field}", apiHistoricalCountMap) // ?days=28
		admin.HandleFunc("/admin/by-day", apiRecordsByDay)                       // ?days=28
		admin.HandleFunc("/admin/installs/{uid}", apiInstallByUid)               // ?days=28
		admin.HandleFunc("/admin/records/{id}", apiRecordById)                   // nothing
		authed := httpauth.SimpleBasicAuth(user, pass)(admin)

		router.Handle("/admin", authed)
		router.Handle("/admin/{_dummy:.*}", authed)
	}

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"authorization"}),
	)(router)

	logged := handlers.LoggingHandler(os.Stdout, cors)

	listen := c.String("listen")
	log.Info("Listening on ", listen)
	log.Fatal(http.ListenAndServe(listen, logged))
	return nil
}

func serverCheck(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pageok"))
}

func abs(i int) int {
	if i < 0 {
		return -1 * i
	}

	return i
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func clamp(i, x, j int) int {
	return max(i, min(x, j))
}

func round(f float64) int {
	return int(f + 0.5)
}

func serverRoot(w http.ResponseWriter, req *http.Request) {
	nRows := 15
	nCols := 80

	var rows [][]byte
	for y := 0; y < nRows; y++ {
		rows = append(rows, make([]byte, nCols+1, nCols+1))

		for x := 0; x <= nCols; x++ {
			if x == 0 || x == nCols-1 {
				rows[y][x] = '|'
			} else if x == nCols {
				rows[y][x] = '\n'
			} else {
				rows[y][x] = ' '
			}
		}
	}

	y := nRows / 2
	ly := y
	dy := 0.0
	for x := 3; x < nCols-2; x++ {
		log.Debugf("y=%d, ly=%d, dy=%f", y, ly, dy)
		rows[y][x] = 'X'
		rows[y][x-1] = 'X'
		diff := abs(ly - y)
		y1 := min(ly, y) + 1
		y2 := max(ly, y)
		if diff > 1 {
			for z := y1; z < y2; z++ {
				rows[z][x] = 'X'
				rows[z][x-1] = 'X'
			}
		}

		dy += float64(rand.Int()%10)/10.0 - 0.5
		if dy < -2.0 {
			dy = -2.0
		} else if dy > 2.0 {
			dy = 2.0
		}

		ly = y
		y = round(float64(y) + dy)
		if y < 0 {
			y = 0
			dy = -dy
		} else if y > nRows-1 {
			y = nRows - 1
			dy = -dy
		}
	}

	w.Write([]byte(fmt.Sprintf("Rancher Telemetry %s\n", version)))
	w.Write([]byte("+" + strings.Repeat("-", nCols-2) + "+\n"))
	w.Write([]byte("|" + strings.Repeat(" ", nCols-2) + "|\n"))

	for _, row := range rows {
		w.Write(row)
	}

	w.Write([]byte("|" + strings.Repeat(" ", nCols-2) + "|\n"))
	w.Write([]byte("+" + strings.Repeat("-", nCols-2) + "+\n"))
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

func apiActiveInstalls(w http.ResponseWriter, req *http.Request) {
	hours, err := getHours(req, 7)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	installs, err := dbPublisher.GetActiveInstalls(hours)
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

func apiLatestCounts(w http.ResponseWriter, req *http.Request) {
	hours, err := getHours(req, 7)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	vars := mux.Vars(req)
	str := vars["fields"]
	fields := strings.Split(str, ",")

	log.Debugf("Fields: %s", fields)
	if len(fields) == 0 {
		respondError(w, req, "You must provide some fields...", 422)
	}

	data, err := dbPublisher.SumOfActiveInstalls(hours, fields)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func apiLatestCountMap(w http.ResponseWriter, req *http.Request) {
	hours, err := getHours(req, 7)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	vars := mux.Vars(req)
	field := vars["field"]
	log.Debugf("Field: %s", field)
	if field == "" {
		respondError(w, req, "You must provide a field...", 422)
	}

	data, err := dbPublisher.SumOfActiveInstallsMap(hours, field)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func apiHistoricalCounts(w http.ResponseWriter, req *http.Request) {
	days, err := getDays(req, 28)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	vars := mux.Vars(req)
	str := vars["fields"]
	fields := strings.Split(str, ",")

	log.Debugf("Fields: %s", fields)
	if len(fields) == 0 {
		respondError(w, req, "You must provide some fields...", 422)
	}

	data, err := dbPublisher.SumByDay(days, fields)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func apiHistoricalCountMap(w http.ResponseWriter, req *http.Request) {
	days, err := getDays(req, 28)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	vars := mux.Vars(req)
	field := vars["field"]
	log.Debugf("Field: %s", field)
	if field == "" {
		respondError(w, req, "You must provide a field...", 422)
	}

	data, err := dbPublisher.SumByDayMap(days, field)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func apiRecordsByDay(w http.ResponseWriter, req *http.Request) {
	days, err := getDays(req, 28)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	data, err := dbPublisher.GetRecordsGroupedByDay(days)
	if err != nil {
		respondError(w, req, err.Error(), 500)
		return
	}

	respondSuccess(w, req, data)
}

func apiInstallByUid(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uid := vars["uid"]
	if uid == "" {
		respondError(w, req, "UID is required", 422)
		return
	}

	days, err := getDays(req, 28)
	if err != nil {
		respondError(w, req, err.Error(), 422)
		return
	}

	records, err := dbPublisher.GetRecordsByUid(uid, days)
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

func apiRecordById(w http.ResponseWriter, req *http.Request) {
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

func getHours(req *http.Request, def int) (int, error) {
	out := def

	str := req.URL.Query().Get("hours")
	if str != "" {
		num, err := strconv.Atoi(str)
		if err == nil {
			out = num
		}
	}

	if out < 1 {
		return 0, errors.New("Hours must be > 0")
	} else {
		return out, nil
	}
}

func getDays(req *http.Request, def int) (int, error) {
	out := def

	str := req.URL.Query().Get("days")
	if str != "" {
		num, err := strconv.Atoi(str)
		if err == nil {
			out = num
		}
	}

	if out < 1 {
		return 0, errors.New("Days must be > 0")
	} else {
		return out, nil
	}
}
