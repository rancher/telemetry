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
		log.Errorf("Error commiting transatcion: %s", err)
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
