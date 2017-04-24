package util

import (
	// "github.com/stuphlabs/pullcord"
	"log/syslog"
	"sync"
)

const syslogFacility = syslog.LOG_DAEMON
const syslogIdentity = "Pullcord"

var syslogger *syslog.Writer

var once sync.Once

func log() *syslog.Writer {
	once.Do(func() {
		var err error
		syslogger, err = syslog.New(syslogFacility, syslogIdentity)
		if err != nil {
			panic(err)
		}
	})
	return syslogger
}

