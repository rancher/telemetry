package publish

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	record "github.com/rancher/telemetry/record"
)

type Postgres struct {
	telemetryVersion string

	Conn *sql.DB
}

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
		log.Errorf("Error creating transaction: %s", err)
		tx.Rollback()
		return err
	}

	recordId, err := p.addRecord(tx, uid, r)
	log.Debugf("Add Record: %v, %s", recordId, err)
	if err != nil {
		log.Errorf("Error adding record: %s", err)
		tx.Rollback()
		return err
	}

	_, err = p.upsertInstall(tx, uid, clientIp, recordId)
	if err != nil {
		log.Errorf("Error updating install: %s", err)
		return err
	}

	_, err = p.upsertByDay(tx, uid, recordId)
	if err != nil {
		log.Errorf("Error updating day: %s", err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("Error commiting transaction: %s", err)
		tx.Rollback()
		return err
	}

	log.Debugf("Published to Postgres")
	return nil
}

func (p *Postgres) LicenseInstallation(r record.Record, clientIp string) (*License, error) {
	log.Debugf("Licensing Installation to Postgres")

	license, ok := r["license"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("  Error getting license installation data")
	}
	licenseKey, ok := license["key"].(string)
	if !ok {
		return nil, fmt.Errorf("  Error getting license installation key")
	}
	installUid, ok := license["installationUid"].(string)
	if !ok {
		return nil, fmt.Errorf("  Error getting license installation installationUid")
	}
	telemetryUid, ok := license["telemetryUid"].(string)
	if !ok {
		return nil, fmt.Errorf("  Error getting license installation telemetryUid")
	}
	nodes, ok := license["runningNodes"].(float64)
	if !ok {
		return nil, fmt.Errorf("  Error getting license installation runningNodes")
	}
	runningNodes := int(nodes)

	tx, err := p.Conn.Begin()
	if err != nil {
		log.Errorf("  Error creating transaction: %s", err)
		tx.Rollback()
		return nil, err
	}

	recordId, err := p.addLicenseIntallationRecord(tx, installUid, licenseKey, telemetryUid, r)
	log.Debugf("  Add License Installation Record: %v, %s", recordId, err)
	if err != nil {
		log.Errorf("    Error adding license installation record: %s", err)
		tx.Rollback()
		return nil, err
	}

	recordId, err = p.upsertLicenseInstallation(tx, installUid, licenseKey, telemetryUid, clientIp, runningNodes, recordId)
	log.Debugf("  Update License Installation: %v, %s", recordId, err)
	if err != nil {
		log.Errorf("    Error updating license installation: %s", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("  Error commiting transaction: %s", err)
		tx.Rollback()
		return nil, err
	}

	log.Debugf("Licensed Installation to Postgres")
	return p.GetLicenseByKey(licenseKey)
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

func (p *Postgres) addLicenseIntallationRecord(tx *sql.Tx, uid, licenseKey, telemetryUid string, r record.Record) (int, error) {
	var id int

	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(`
		INSERT INTO license_installation_record(uid,license_key,telemetry_uid,data,ts) 
		VALUES ($1,$2,$3,$4,NOW())
		RETURNING id`, uid, licenseKey, telemetryUid, string(b)).Scan(&id)
	return id, err
}

func (p *Postgres) addLicenseRecord(tx *sql.Tx, key string, r record.Record) (int, error) {
	var id int

	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(`
		INSERT INTO license_record(key,data,ts) 
		VALUES ($1,$2,NOW())
		RETURNING id`, key, string(b)).Scan(&id)
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

func (p *Postgres) upsertLicenseInstallation(tx *sql.Tx, uid, licenseKey, telemetryUid, clientIp string, runningNodes, recordId int) (int, error) {
	var id int

	err := tx.QueryRow(`
		INSERT INTO license_installation(uid,license_key,telemetry_uid,last_ip,running_nodes,last_record,first_seen,last_seen) 
		VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())
		ON CONFLICT(uid) DO UPDATE SET
			last_seen=NOW(),
			license_key=$2,
			telemetry_uid=$3,
			last_ip=$4,
			running_nodes=$5,
			last_record=$6
		RETURNING id`, uid, licenseKey, telemetryUid, clientIp, runningNodes, recordId).Scan(&id)
	return id, err
}

func (p *Postgres) upsertLicense(tx *sql.Tx, key, clientIp string, lInstallations, lNodes, recordId int) (int, error) {
	var id int

	err := tx.QueryRow(`
		INSERT INTO license(key,last_ip,licensed_installations,licensed_nodes,last_record,first_seen,last_seen) 
		VALUES ($1,$2,$3,$4,$5,NOW(),NOW())
		ON CONFLICT(uid) DO UPDATE SET
			last_seen=NOW(),
			last_ip=$2,
			licensed_installations=$3,
			licensed_nodes=$4,
			last_record=$5
		RETURNING id`, key, clientIp, lInstallations, lNodes, recordId).Scan(&id)
	return id, err
}

func (p *Postgres) upsertByDay(tx *sql.Tx, uid string, recordId int) (int, error) {
	var id int

	today := time.Now().Format("2006-01-02")

	err := tx.QueryRow(`
		INSERT INTO byday(uid,day,record_id)
		VALUES ($1,$2,$3) 
		ON CONFLICT(uid,day) DO UPDATE SET 
			record_id=$3
		RETURNING id`, uid, today, recordId).Scan(&id)
	return id, err
}

func (p *Postgres) GetAccountHash(user string) (string, error) {
	var hash string
	err := p.Conn.QueryRow(`SELECT hash FROM account WHERE name=$1`, user).Scan(&hash)
	if err != nil {
		return "", err
	}

	return hash, nil
}
