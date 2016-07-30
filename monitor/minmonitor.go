package monitor

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	// "github.com/stuphlabs/pullcord"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const minSessionCookieNameRandSize = 32
const minSessionCookieValueRandSize = 128
const minSessionCookieMaxAge = 2 * 60 * 60

type monitorredService struct {
	address string
	port uint16
	protocol string
	gracePeriod uint32
	lastChecked time.Time
	up bool
}

type MinMonitor struct {
	table map[string]monitorredService
}

func (monitor *MinMonitor) Add(
	name string,
	address string,
	port string,
	protocol string,
	gracePeriod uint32,
	deferProbe bool,
) (err error) {
	osvc, previousEntryExists := monitor.table[name]
	if previousEntryExists {
		log().Error(
			fmt.Sprintf(
				"minmonitor cannot add service: name \"%s\"" +
				" previously used for service at protocol" +
				" \"%s\" address \"%s\" port \"%d\", new" +
				" service at protocol \"%s\" address \"%s\"" +
				" port \"%d\"",
				name,
				osvc.protocol,
				osvc.address,
				osvc.port,
				protocol,
				address,
				port,
			),
		)

		return //TODO
	}

	monitor.table[name] = monitorredService{
		address: address,
		port: port,
		protocol: protocol,
		gracePeriod: gracePeriod,
		lastChecked: 0,
		up: false,
	}

	if ! deferProbe {
		_, err := monitor.Reprobe(name)

		if err != nil {
			log().Warning(
				fmt.Sprintf(
					"minmonitor had an error during the" +
					" initial probe for service \"%s\": %v",
					name,
					err,
				),
			)
		}
	}

	log().Info(
		fmt.Sprintf(
			"minmonitor has successfully added service: \"%s\"",
			name,
		)
	)
	return err
}

func (monitor *MinMonitor) Reprobe(name string) (up bool, err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Error(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:" +
				" \"%s\"",
				name,
			),
		)

		return false, //TODO
	}

	conn, err := net.Dial(svc.protocol, svc.address + ":" + svc.port)
	svc.lastChecked = time.Now()
	if err != nil {
		svc.up = false
		// TODO check what the error was

		log().Warning(
			fmt.Sprintf(
				"minmonitor encountered an error while" +
				" probing service \"%s\": %v",
				name,
				err,
			),
		)

		return false, err
	} else {
		defer conn.Close()
		svc.up = true

		log().Info(
			fmt.Sprintf(
				"minmonitor successfully probed service:" +
				" \"%s\"",
				name,
			),
		)

		return true, nil
	}
}

func (monitor *MinMonitor) Status(name string) (up bool, err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Error(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:" +
				" \"%s\"",
				name,
			),
		)

		return false, //TODO
	}

	if (! svc.up) || (time.Now() > svc.lastChecked + svc.gracePeriod) {
		return monitor.Reprobe(name)
	} else {
		log().Info(
			fmt.Sprintf(
				"minmonitor is skipping the reprobe as the" +
				" current time is still within the grace" +
				" period of the last successfull probe of" +
				" service: \"%s\"",
				name,
			),
		)

		return true, nil
	}
}

func (monitor *MinMonitor) SetStatusUp(name string) (err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Error(
			fmt.Sprintf(
				"minmonitor cannot set the status of unknown" +
				" service: \"%s\"",
				name,
			),
		)

		return foo
	}

	log().Info(
		fmt.Sprintf(
			"minmonitor has been explicitly informed of the up" +
			" status of service: \"%s\"",
			name,
		),
	)
	svc.lastChecked = time.Now()
	svc.up = true
}

// NewMinSessionHandler constructs a new MinSessionHandler given a unique name
// (which will be given to all the cookies), and a path and domain (the two of
// which will simply be sent to the browser along with the cookie, and otherwise
// have no bearing on functionality).
func NewMinMonitor() *MinMonitor {
	log().Info("initializing minimal service monitor")

	var result MinMonitor

	return &result
}
