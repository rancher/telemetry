package publish

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/urfave/cli"

	record "github.com/rancher/telemetry/record"
)

type Postgres struct {
	telemetryVersion string

	Conn *sql.DB
}

type ApiInstallation struct {
	Id        int64       `json:"id"`
	Uid       string      `json:"uid"`
	FirstSeen time.Time   `json:"first_seen"`
	LastSeen  time.Time   `json:"last_seen"`
	LastIp    string      `json:"last_ip"`
	Record    interface{} `json:"record"`
}

type ApiRecord struct {
	Id     int64       `json:"id"`
	Uid    string      `json:"uid"`
	Ts     time.Time   `json:"ts"`
	Record interface{} `json:"record"`
}

type RecordsByUid map[string]ApiRecord
type RecordsByDateByUid map[string]RecordsByUid

type AggregatedFields map[string]int64
type AggregatedFieldsByDate map[string]AggregatedFields

func NewPostgres(c *cli.Context) *Postgres {
	host := c.String("pg-host")
	port := c.String("pg-port")
	user := c.String("pg-user")
	pass := c.String("pg-pass")
	dbname := c.String("pg-dbname")
	sslmode := c.String("pg-ssl")

	out := &Postgres{
		telemetryVersion: c.App.Version,
	}

	if host != "" && user != "" && pass != "" {
		log.Info("Postgres enabled")
	} else {
		return out
	}

	dsn := strings.Join([]string{
		"host=" + host,
		"port=" + port,
		"user=" + user,
		"password='" + pass + "'",
		"dbname=" + dbname,
		"sslmode=" + sslmode,
	}, " ")

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error connecting to DB: %s", err)
	}

	out.Conn = conn
	err = out.testDb()
	if err != nil {
		log.Fatalf("Error connecting to DB: %s", err)
	}

	log.Infof("Connected to Postgres at %s", host)
	return out
}

func (p *Postgres) Report(r record.Record, clientIp string) error {
	log.Debugf("Publishing to Postgres")

	install := r["install"].(map[string]interface{})
	uid := install["uid"].(string)

	tx, err := p.Conn.Begin()
	if err != nil {
		log.Debugf("Error creating transaction: %s", err)
		tx.Rollback()
		return err
	}

	recordId, err := p.addRecord(tx, uid, r)
	log.Debugf("Add Record: %s, %s", recordId, err)
	if err != nil {
		log.Debugf("Error adding record: %s", err)
		tx.Rollback()
		return err
	}

	_, err = p.upsertInstall(tx, uid, clientIp, recordId)
	if err != nil {
		log.Debugf("Error updating install: %s", err)
		return err
	}

	_, err = p.upsertByDay(tx, uid, recordId)
	if err != nil {
		log.Debugf("Error updating day: %s", err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Debugf("Error commiting transatgion: %s", err)
		tx.Rollback()
		return err
	}

	log.Debugf("Published to Postgres")
	return nil
}

func (p *Postgres) testDb() error {
	var one int
	err := p.Conn.QueryRow(`SELECT 1`).Scan(&one)
	if err != nil {
		return err
	}

	if one != 1 {
		return errors.New(fmt.Sprintf("SELECT 1 == %d?!", one))
	}

	return nil
}

func (p *Postgres) addRecord(tx *sql.Tx, uid string, r record.Record) (int, error) {
	var id int

	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(`INSERT INTO record(uid,data,ts) VALUES ($1,$2,NOW()) RETURNING id`, uid, string(b)).Scan(&id)
	return id, err
}

func (p *Postgres) upsertInstall(tx *sql.Tx, uid string, clientIp string, recordId int) (int, error) {
	var id int

	err := tx.QueryRow(`
		INSERT INTO installation(uid,last_ip,last_record,first_seen,last_seen)
		VALUES ($1,$2,$3,NOW(),NOW()) 
		ON CONFLICT(uid) DO UPDATE SET 
			last_seen=NOW(),
			last_ip=$2,
			last_record=$3
		RETURNING id`, uid, clientIp, recordId).Scan(&id)
	return id, err
}

func (p *Postgres) upsertByDay(tx *sql.Tx, uid string, recordId int) (int, error) {
	var id int

	err := tx.QueryRow(`
		INSERT INTO byday(uid,ts,record_id)
		VALUES ($1,$2,$3) 
		ON CONFLICT(uid,day) DO UPDATE SET 
			record_id=$3
		RETURNING id`, uid, recordId).Scan(&id)
	return id, err
}

// ----------------------
// Queries for Admin API
// ----------------------
func (p *Postgres) GetActiveInstalls(hours int) ([]ApiInstallation, error) {
	sql := `
		SELECT i.id, i.uid, i.first_seen, i.last_seen, i.last_ip, r.data
		FROM installation i
			JOIN record r ON (i.last_record = r.id)
		WHERE i.last_seen >= NOW() - INTERVAL '%d hour'`

	rows, err := p.Conn.Query(fmt.Sprintf(sql, hours))

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	out := []ApiInstallation{}

	defer rows.Close()
	for rows.Next() {
		var i ApiInstallation
		var data []byte
		err = rows.Scan(&i.Id, &i.Uid, &i.FirstSeen, &i.LastSeen, &i.LastIp, &data)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &i.Record)
		if err != nil {
			return nil, err
		}

		out = append(out, i)
	}

	return out, nil
}

func (p *Postgres) GetRecordsGroupedByDay(days int) (RecordsByDateByUid, error) {
	sql := `
		SELECT id, uid, ts, data
		FROM record
		WHERE date_trunc('day',ts) >= (date_trunc('day',now()) - INTERVAL '%d day')
		ORDER BY id DESC`

	rows, err := p.Conn.Query(fmt.Sprintf(sql, days))

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	out := make(RecordsByDateByUid)

	defer rows.Close()
	for rows.Next() {
		var rec ApiRecord
		var data []byte
		err = rows.Scan(&rec.Id, &rec.Uid, &rec.Ts, &data)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &rec.Record)
		if err != nil {
			return nil, err
		}

		day := rec.Ts.Format("2006-01-02")
		byDate, ok := out[day]
		if !ok {
			byDate = make(RecordsByUid)
			out[day] = byDate
		}

		_, exists := byDate[rec.Uid]
		if !exists {
			byDate[rec.Uid] = rec
		}
	}

	return out, nil
}

func (p *Postgres) GetRecordsByUid(uid string, days int) ([]ApiRecord, error) {
	sql := `
		SELECT id, uid, ts, data
		FROM record
		WHERE 
			uid = $1
		  AND date_trunc('day',ts) >= (date_trunc('day',now()) - INTERVAL '%d day')`

	rows, err := p.Conn.Query(fmt.Sprintf(sql, days), uid)

	if err != nil {
		return nil, err
	}

	out := []ApiRecord{}

	defer rows.Close()
	for rows.Next() {
		var rec ApiRecord
		var data []byte
		err = rows.Scan(&rec.Id, &rec.Uid, &rec.Ts, &data)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &rec.Record)
		if err != nil {
			return nil, err
		}

		out = append(out, rec)
	}

	return out, nil
}

func (p *Postgres) GetRecordById(id string) (ApiRecord, error) {
	sql := `
		SELECT id, uid, ts, data
		FROM record
		WHERE 
			id = $1`

	var rec ApiRecord
	var data []byte

	err := p.Conn.QueryRow(sql, id).Scan(&rec.Id, &rec.Uid, &rec.Ts, &data)
	if err != nil {
		return rec, err
	}

	err = json.Unmarshal(data, &rec.Record)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

func (p *Postgres) SumOfActiveInstalls(hours int, fields []string) (AggregatedFields, error) {
	fieldSql, err := fieldQuery(fields, "r.data")
	if err != nil {
		return nil, err
	}

	sql := `SELECT
	%s
FROM installation i
	JOIN record r ON (i.last_record = r.id)
WHERE i.last_seen >= NOW() - INTERVAL '%d hour'`

	sql = fmt.Sprintf(sql, fieldSql, hours)

	log.Debugf("SQL: %s", sql)

	rows, err := p.Conn.Query(sql)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := make(AggregatedFields)
	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}

	rows.Next()
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(cols); i++ {
		out[fields[i]] = (*(vals[i].(*interface{}))).(int64)
	}

	return out, nil
}

func (p *Postgres) SumByDay(days int, fields []string) (AggregatedFieldsByDate, error) {
	sql := `SELECT
	%s,
	b.day
FROM byday b
	JOIN record r on (b.record_id=r.id)
WHERE b.day >= (to_date('%s','YYYY-MM-DD') - INTERVAL '%d day')
GROUP BY day
ORDER BY day`

	today := time.Now().Format("2006-01-02")

	fieldSql, err := fieldQuery(fields, "r.data")
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(sql, fieldSql, today, days)

	log.Debugf("SQL: %s", sql)

	rows, err := p.Conn.Query(sql)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := make(AggregatedFieldsByDate)

	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, err
		}

		entry := make(AggregatedFields)

		for i := 0; i < len(cols)-1; i++ {
			switch val := (*(vals[i].(*interface{}))).(type) {
			case int64:
				entry[fields[i]] = val
			}
		}

		day := (*(vals[len(cols)-1].(*interface{}))).(time.Time)
		dayStr := day.Format("2006-01-02")

		out[dayStr] = entry
	}

	return out, nil
}

func fieldQuery(fields []string, dataField string) (string, error) {
	validField := regexp.MustCompile("^[a-zA-Z0-9._-]+$")

	out := []string{}

	for _, field := range fields {
		if !validField.MatchString(field) {
			return "", errors.New("Invalid field")
		}

		parts := strings.Split(field, ".")
		prefix := ""
		fn := ""
		suffix := ""
		if strings.HasSuffix(field, "_min") {
			fn = "min"
		} else if strings.HasSuffix(field, "_avg") {
			prefix = "round("
			fn = "avg"
			suffix = ")::int"
		} else if strings.HasSuffix(field, "_max") {
			fn = "max"
		} else {
			fn = "sum"
		}

		out = append(out, prefix+fn+"(json_extract_path("+dataField+",'"+strings.Join(parts, "','")+"')::text::int)"+suffix+" AS \""+field+"\"")
	}

	return strings.Join(out, ",\n"), nil
}
