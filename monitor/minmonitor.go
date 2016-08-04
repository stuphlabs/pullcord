package monitor

import (
	"fmt"
	// "github.com/stuphlabs/pullcord"
	"net"
	"strconv"
	"time"
)

// errorString is an unexported error type used for error constants.
type errorString string

// Error allows the unexported error type used for some error constants in this
// package to adhere to the standard error interface.
func (str errorString) Error() string {
	return string(str)
}

// DuplicateServiceRagistrationError indicates that a service with that name
// has already been registered to this monitor.
const DuplicateServiceRagistrationError = errorString(
	"A service with this name has already been registered",
)

// UnknownServiceError indicates that no service with the given name has been
// registered.
const UnknownServiceError = errorString(
	"No service has been registered with the requested name",
)

// monitorredService holds the information for a single service definition.
type monitorredService struct {
	address string
	port uint16
	protocol string
	gracePeriod time.Duration
	lastChecked time.Time
	up bool
}

// MinMonitor is a minimal service monitor not intended to be used in
// production. Named services will have an up status cached for a time, while a
// down status will never be cached. It is possible to explicitly set a service
// as being up (which will be cached as with a normal probe). It is also
// possible to explicitly re-probe a service regardless of the status of the
// cache.
type MinMonitor struct {
	table map[string]monitorredService
}

// Add adds a named service to the monitor. The named service is associated
// with a specific protocol (i.e. TCP, UDP) on a specific port at a specific
// address. A grace period can be used to keep this monitor from unnecessarily
// overwhelming the service by allowing an up status to be cached for a time.
// This function also allows the initial probe to either begin immediately or
// to be deferred until the first status request.
//
// At this time, it is suggested that only TCP services be probed as the
// current version of MinMonitor only checks to see that net.Dial() does not
// fail, which would not be enough information to make a determination of the
// status of a service that communicates over UDP. As the inability to interact
// beyond an attempt to open a connection is a handicap in determining even the
// status of some TCP-based services, future monitor implementations (including
// any intended to be used in a production environment) should allow service
// status to be determined by some amount of specified interaction (perhaps
// involving regular expressions or callback functions).
func (monitor *MinMonitor) Add(
	name string,
	address string,
	port uint16,
	protocol string,
	gracePeriod time.Duration,
	deferProbe bool,
) (err error) {
	osvc, previousEntryExists := monitor.table[name]
	if previousEntryExists {
		log().Err(
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

		return DuplicateServiceRagistrationError
	}

	monitor.table[name] = monitorredService{
		address: address,
		port: port,
		protocol: protocol,
		gracePeriod: gracePeriod,
		lastChecked: time.Time{},
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
		),
	)
	return err
}

// Reprobe forces the status of the named service to be checked immediately
// without regard to a possible previously cached up status. The result of this
// probe will automatically be cached by the monitor.
func (monitor *MinMonitor) Reprobe(name string) (up bool, err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Err(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:" +
				" \"%s\"",
				name,
			),
		)

		return false, UnknownServiceError
	}

	conn, err := net.Dial(
		svc.protocol,
		svc.address + ":" + strconv.Itoa(int(svc.port)),
	)
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

// Status returns true if the status of the named service is currently believed
// to be up. The service could have its status reported as being up if a brand
// new probe of the service indicates that the service is indeed up, or if a
// recent probe indicated that the service was up (specifically if the most
// recent probe indicated that the service was up and that probe was within a
// grace period that was specified when the service was registered), or if the
// status of the service was explicitly set as being up within that same grace
// period (and no other forced re-probe has occurred since this forced status
// up assignment). However, if the status of the service is reported as being
// down, then it necessarily means that a probe has just occurred and the
// service was unable to be reached.
func (monitor *MinMonitor) Status(name string) (up bool, err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Err(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:" +
				" \"%s\"",
				name,
			),
		)

		return false, UnknownServiceError
	}

	if (! svc.up) || time.Now().After(
		svc.lastChecked.Add(svc.gracePeriod),
	) {
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

// SetStatusUp explicitly sets the status of a named service as being up. This
// up status will be cached just as if it were the result of a normal probe.
func (monitor *MinMonitor) SetStatusUp(name string) (err error) {
	svc, entryExists := monitor.table[name]
	if ! entryExists {
		log().Err(
			fmt.Sprintf(
				"minmonitor cannot set the status of unknown" +
				" service: \"%s\"",
				name,
			),
		)

		return UnknownServiceError
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

	return nil
}

// NewMinMonitor constructs a new MinMonitor.
func NewMinMonitor() *MinMonitor {
	log().Info("initializing minimal service monitor")

	var result MinMonitor
	result.table = make(map[string]monitorredService)

	return &result
}
