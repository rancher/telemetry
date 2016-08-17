package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	cmd "github.com/rancher/telemetry/cmd"
)

var (
	VERSION string
)

func main() {
	app := cli.NewApp()
	app.Name = "telemetry"
	app.Author = "Rancher Labs, Inc."
	app.Usage = "Rancher telemetry"

	if VERSION == "" {
		app.Version = "git"
	} else {
		app.Version = VERSION
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug logging",
			EnvVar: "TELEMETRY_DEBUG",
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
	}
	app.Before = before
	app.Commands = []cli.Command{
		cmd.ClientCommand(),
		cmd.ServerCommand(),
	}

	app.Run(os.Args)
}

func before(c *cli.Context) error {
	timeFormat := new(log.TextFormatter)
	timeFormat.TimestampFormat = "2006-01-02 15:04:05"
	timeFormat.FullTimestamp = true
	log.SetFormatter(timeFormat)

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

	return nil
}
