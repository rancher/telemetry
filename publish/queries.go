package publish

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

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

type LicenseIntallation struct {
	Uid          string `json:"uid"`
	TelemetryUID string `json:"telemetryUid"`
	RunningNodes int    `json:"runningNodes"`
	LastIp       string `json:"lastIp"`
}

type License struct {
	Key                  string               `json:"key"`
	Intallations         []LicenseIntallation `json:"installations"`
	LicensedIntallations int                  `json:"licensedIntallations"`
	LicensedNodes        int                  `json:"licensedNodes"`
	RunningIntallations  int                  `json:"runningIntallations"`
	RunningNodes         int                  `json:"runningNodes"`
	Compliant            bool                 `json:"compliant"`
}

type RecordsByUid map[string]ApiRecord
type RecordsByDateByUid map[string]RecordsByUid

type AggregatedFields map[string]int64
type AggregatedFieldsByDate map[string]AggregatedFields

func (p *Postgres) Ping() error {
	sql := `SELECT 1`
	var one int
	err := p.Conn.QueryRow(sql).Scan(&one)
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgres) GetAllInstalls() ([]ApiInstallation, error) {
	sql := `SELECT id, uid, first_seen, last_seen, i.last_ip FROM installation i ORDER BY first_seen`
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ApiInstallation{}

	for rows.Next() {
		var i ApiInstallation
		err = rows.Scan(&i.Id, &i.Uid, &i.FirstSeen, &i.LastSeen, &i.LastIp)
		if err != nil {
			return nil, err
		}

		out = append(out, i)
	}

	return out, nil
}

func (p *Postgres) GetLicenseByKey(key string) (*License, error) {
	sql := `SELECT licensed_installations,licensed_nodes FROM license WHERE key = $1`

	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &License{Key: key}

	rows.Next()
	err = rows.Scan(&out.LicensedIntallations, &out.LicensedNodes)
	if err != nil {
		return nil, err
	}

	out.RunningNodes, out.Intallations, err = p.GetLicenseInstallationByLicenseKey(key)
	if err != nil {
		return nil, err
	}

	out.RunningIntallations = len(out.Intallations)
	out.Compliant = (out.LicensedNodes >= out.RunningNodes) && (out.LicensedIntallations >= out.RunningIntallations)

	return out, nil
}

func (p *Postgres) GetLicenseInstallationByLicenseKey(licenseKey string) (int, []LicenseIntallation, error) {
	sql := `SELECT uid, telemetry_uid, last_ip, running_nodes FROM license_installation WHERE license_key = $1`

	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, licenseKey)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	out := []LicenseIntallation{}
	totalRunningNodes := 0
	for rows.Next() {
		var i LicenseIntallation

		err = rows.Scan(&i.Uid, &i.TelemetryUID, &i.LastIp, &i.RunningNodes)
		if err != nil {
			return 0, nil, err
		}

		totalRunningNodes += i.RunningNodes

		out = append(out, i)
	}

	return totalRunningNodes, out, nil
}

func (p *Postgres) GetActiveInstalls(hours int) ([]ApiInstallation, error) {
	sql := `SELECT i.id, i.uid, i.first_seen, i.last_seen, i.last_ip, r.data
FROM installation i
	JOIN record r ON (i.last_record = r.id)
WHERE i.last_seen >= NOW() - INTERVAL '%d hour'`

	sql = fmt.Sprintf(sql, hours)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ApiInstallation{}

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

func (p *Postgres) GetActiveCountByDay() (AggregatedFields, error) {
	sql := `SELECT date_trunc('day',ts) AS day, count(DISTINCT uid) FROM record GROUP BY day ORDER BY DAY`
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(AggregatedFields)

	for rows.Next() {
		var date time.Time
		var count int64

		err = rows.Scan(&date, &count)
		if err != nil {
			return nil, err
		}

		day := date.Format("2006-01-02")
		out[day] = count
	}

	return out, nil
}

func (p *Postgres) GetRecordsGroupedByDay(days int) (RecordsByDateByUid, error) {
	sql := `SELECT id, uid, ts, data
FROM record
WHERE date_trunc('day',ts) >= (date_trunc('day',now()) - INTERVAL '%d day')
ORDER BY id DESC`

	sql = fmt.Sprintf(sql, days)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(RecordsByDateByUid)

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
	sql := `SELECT id, uid, ts, data
FROM record
WHERE 
	uid = $1
	AND date_trunc('day',ts) >= (date_trunc('day',now()) - INTERVAL '%d day')
ORDER BY id DESC`

	sql = fmt.Sprintf(sql, days)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ApiRecord{}

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
	sql := `SELECT id, uid, ts, data
FROM record
WHERE 
	id = $1`

	var rec ApiRecord
	var data []byte

	log.Debugf("Query: %s", sql)
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
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (p *Postgres) SumOfActiveInstallsMap(hours int, field string) (AggregatedFields, error) {
	if !fieldIsValid(field) {
		return nil, errors.New("Invalid field")
	}

	parts := strings.Split(field, ".")
	path := "'" + strings.Join(parts, "','") + "'"

	sql := `SELECT jet.key, sum(jet.value::int)
FROM installation i
	JOIN record r ON (i.last_record = r.id),
	json_each_text(json_extract_path(r.data,%s)) AS jet
WHERE i.last_seen >= NOW() - INTERVAL '%d hour'
GROUP BY jet.key
ORDER BY jet.key`

	sql = fmt.Sprintf(sql, path, hours)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(AggregatedFields)

	for rows.Next() {
		var key string
		var val int64

		err = rows.Scan(&key, &val)
		if err != nil {
			return nil, err
		}

		out[key] = val
	}

	return out, nil
}

func (p *Postgres) SumOfActiveInstallsValue(hours int, field string) (AggregatedFields, error) {
	if !fieldIsValid(field) {
		return nil, errors.New("Invalid field")
	}

	parts := strings.Split(field, ".")
	path := "'" + strings.Join(parts, "','") + "'"

	sql := `SELECT key, count(*) AS value 
FROM installation i
	JOIN record r ON (i.last_record = r.id),
	json_extract_path_text(r.data,%s) AS key
WHERE i.last_seen >= NOW() - INTERVAL '%d hour'
GROUP BY key
ORDER BY value DESC`

	sql = fmt.Sprintf(sql, path, hours)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(AggregatedFields)

	for rows.Next() {
		var key string
		var val int64

		err = rows.Scan(&key, &val)
		if err != nil {
			return nil, err
		}

		out[key] = val
	}

	return out, nil
}

func (p *Postgres) SumByDay(days int, fields []string, uid string) (AggregatedFieldsByDate, error) {
	sql := `SELECT
	%s,
	b.day
FROM byday b
	JOIN record r on (b.record_id=r.id)
WHERE b.day >= (to_date('%s','YYYY-MM-DD') - INTERVAL '%d day')
	AND b.uid %s $1
GROUP BY day
ORDER BY day`

	today := time.Now().Format("2006-01-02")

	fieldSql, err := fieldQuery(fields, "r.data")
	if err != nil {
		return nil, err
	}

	op := "="
	if uid == "" {
		op = "<>"
	}

	sql = fmt.Sprintf(sql, fieldSql, today, days, op)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (p *Postgres) SumByDayMap(days int, field string, uid string) (AggregatedFieldsByDate, error) {
	if !fieldIsValid(field) {
		return nil, errors.New("Invalid field")
	}

	today := time.Now().Format("2006-01-02")

	parts := strings.Split(field, ".")
	path := "'" + strings.Join(parts, "','") + "'"

	sql := `SELECT b.day, jet.key, sum(jet.value::int)
FROM byday b
	JOIN record r ON (b.record_id = r.id),
	json_each_text(json_extract_path(r.data,%s)) AS jet
WHERE b.day >= (to_date('%s','YYYY-MM-DD') - INTERVAL '%d day')
	AND b.uid %s $1
GROUP BY b.day, jet.key
ORDER BY b.day, jet.key`

	op := "="
	if uid == "" {
		op = "<>"
	}

	sql = fmt.Sprintf(sql, path, today, days, op)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(AggregatedFieldsByDate)

	for rows.Next() {
		var day time.Time
		var key string
		var val int64

		err = rows.Scan(&day, &key, &val)

		if err != nil {
			return nil, err
		}

		dayStr := day.Format("2006-01-02")
		byDate, ok := out[dayStr]
		if !ok {
			byDate = make(AggregatedFields)
			out[dayStr] = byDate
		}

		byDate[key] = val
	}

	return out, nil
}

func (p *Postgres) SumByDayValue(days int, field string, uid string) (AggregatedFieldsByDate, error) {
	if !fieldIsValid(field) {
		return nil, errors.New("Invalid field")
	}

	today := time.Now().Format("2006-01-02")

	parts := strings.Split(field, ".")
	path := "'" + strings.Join(parts, "','") + "'"

	sql := `SELECT b.day, key, count(*) AS value 
FROM byday b
	JOIN record r ON (b.record_id = r.id),
	json_extract_path_text(r.data,%s) AS key
WHERE b.day >= (to_date('%s','YYYY-MM-DD') - INTERVAL '%d day')
	AND b.uid %s $1
GROUP BY b.day, key
ORDER BY b.day, value DESC`

	op := "="
	if uid == "" {
		op = "<>"
	}

	sql = fmt.Sprintf(sql, path, today, days, op)
	log.Debugf("Query: %s", sql)
	rows, err := p.Conn.Query(sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(AggregatedFieldsByDate)

	for rows.Next() {
		var day time.Time
		var key string
		var val int64

		err = rows.Scan(&day, &key, &val)

		if err != nil {
			return nil, err
		}

		dayStr := day.Format("2006-01-02")
		byDate, ok := out[dayStr]
		if !ok {
			byDate = make(AggregatedFields)
			out[dayStr] = byDate
		}

		byDate[key] = val
	}

	return out, nil
}

func fieldIsValid(field string) bool {
	validField := regexp.MustCompile("^[a-zA-Z0-9._-]+$")
	return field != "" && validField.MatchString(field)
}

func fieldQuery(fields []string, dataField string) (string, error) {
	out := []string{}

	for _, field := range fields {
		if !fieldIsValid(field) {
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

		out = append(out, "  "+prefix+fn+"(json_extract_path("+dataField+",'"+strings.Join(parts, "','")+"')::text::int)"+suffix+" AS \""+field+"\"")
	}

	return strings.Join(out, ",\n"), nil
}
