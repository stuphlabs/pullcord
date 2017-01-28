package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	// "github.com/stuphlabs/pullcord"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/proxy"
	"github.com/stuphlabs/pullcord/trigger"
	"net"
	"net/http"
	"strconv"
	"time"
)

// DuplicateServiceRegistrationError indicates that a service with that name
// has already been registered to this monitor.
const DuplicateServiceRegistrationError = errors.New(
	"A service with this name has already been registered",
)

// UnknownServiceError indicates that no service with the given name has been
// registered.
const UnknownServiceError = errors.New(
	"No service has been registered with the requested name",
)

// MonitorredService holds the information for a single service definition.
type MinMonitorredService struct {
	Address string
	Port int
	Protocol string
	GracePeriod time.Duration
	OnDown trigger.TriggerHandler
	OnUp trigger.TriggerHandler
	Always trigger.TriggerHandler
	lastChecked time.Time
	up bool
	passthru falcore.RequestFilter
}

func init() {
	config.RegisterResourceType(
		"minmonitorredservice",
		func() json.Unmarshaler {
			return new(MinMonitorredService)
		},
	)
}

func (s *MinMonitorredService) UnmarshalJSON(data []byte) error {
	var t struct {
		Address string
		Port int
		Protocol string
		GracePeriod string
		OnDown *config.Resource
		OnUp *config.Resource
		Always *config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	if g, e := time.ParseDuration(t.GracePeriod); e != nil {
		return e
	} else {
		s.GracePeriod = g
	}

	if t.OnDown != nil {
		d := t.OnDown.Unmarshaled
		switch d := d.(type) {
		case trigger.TriggerHandler:
			s.OnDown = d
		default:
			return config.UnexpectedResourceType
		}
	} else {
		// TODO test null values for these as well
		s.OnDown = nil
	}

	if t.OnUp != nil {
		u := t.OnUp.Unmarshaled
		switch u := u.(type) {
		case trigger.TriggerHandler:
			s.OnUp = u
		default:
			return config.UnexpectedResourceType
		}
	} else {
		s.OnUp = nil
	}

	if t.Always != nil {
		a := t.Always.Unmarshaled
		switch a := a.(type) {
		case trigger.TriggerHandler:
			s.OnDown = a
		default:
			return config.UnexpectedResourceType
		}
	} else {
		s.Always = nil
	}

	s.Address = t.Address
	s.Port = t.Port
	s.Protocol = t.Protocol

	return nil
}

func NewMinMonitorredService(
	address string,
	port int,
	protocol string,
	gracePeriod time.Duration,
	onDown trigger.TriggerHandler,
	onUp trigger.TriggerHandler,
	always trigger.TriggerHandler,
) (service *MinMonitorredService, err error) {
	result := MinMonitorredService{
		address,
		port,
		protocol,
		gracePeriod,
		onDown,
		onUp,
		always,
		time.Time{},
		false,
		proxy.NewPassthruFilter(address, port),
	}

	return &result, nil
}

// MinMonitor is a minimal service monitor not intended to be used in
// production. Named services will have an up status cached for a time, while a
// down status will never be cached. It is possible to explicitly set a service
// as being up (which will be cached as with a normal probe). It is also
// possible to explicitly re-probe a service regardless of the status of the
// cache.
type MinMonitor struct {
	table map[string]*MinMonitorredService
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
	service *MinMonitorredService,
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
				osvc.Protocol,
				osvc.Address,
				osvc.Port,
				service.Protocol,
				service.Address,
				service.Port,
			),
		)

		return DuplicateServiceRegistrationError
	}

	if monitor.table == nil {
		monitor.table = make(map[string]*MinMonitorredService)
	}

	monitor.table[name] = service

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

	return svc.Reprobe()
}

func (svc *MinMonitorredService) Reprobe() (up bool, err error) {
	conn, err := net.Dial(
		svc.Protocol,
		svc.Address + ":" + strconv.Itoa(int(svc.Port)),
	)
	svc.lastChecked = time.Now()
	if err != nil {
		svc.up = false
		// TODO check what the error was

		switch castErr := err.(type) {
		case *net.OpError:
			if castErr.Addr != nil {
				log().Info(
					fmt.Sprintf(
						"minmonitor received a" +
						" connection refused" +
						" (interpereted as a down" +
						" status) from \"%s:%d\"",
						svc.Address,
						svc.Port,
					),
				)

				return false, nil
			} else {
				log().Warning(
					fmt.Sprintf(
						"minmonitor encountered an" +
						" error while probing" +
						" \"%s:%d\": %v",
						svc.Address,
						svc.Port,
						err,
					),
				)

				return false, err
			}
		default:
			log().Warning(
				fmt.Sprintf(
					"minmonitor encountered an unknown" +
					" error while probing \"%s:%d\": %v",
					svc.Address,
					svc.Port,
					err,
				),
			)

			return false, err
		}
	} else {
		defer conn.Close()
		svc.up = true

		log().Info(
			fmt.Sprintf(
				"minmonitor successfully probed: \"%s:%d\"",
				svc.Address,
				svc.Port,
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

	return svc.Status()
}

func (svc *MinMonitorredService) Status() (up bool, err error) {
	if (! svc.up) || time.Now().After(
		svc.lastChecked.Add(svc.GracePeriod),
	) {
		log().Info(
			fmt.Sprintf(
				"minmonitor must reprobe as either the grace" +
				" period has lapsed or the previous probe" +
				" indicated a down status for: \"%s:%d\"",
				svc.Address,
				svc.Port,
			),
		)

		return svc.Reprobe()
	} else {
		log().Info(
			fmt.Sprintf(
				"minmonitor is skipping the reprobe as the" +
				" current time is still within the grace" +
				" period of the last successfull probe of:" +
				" \"%s:%d\"",
				svc.Address,
				svc.Port,
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

	return svc.SetStatusUp()
}

func (svc *MinMonitorredService) SetStatusUp() (error) {
	log().Info(
		fmt.Sprintf(
			"minmonitor has been explicitly informed of the up" +
			" status of: \"%s:%d\"",
			svc.Address,
			svc.Port,
		),
	)
	svc.lastChecked = time.Now()
	svc.up = true

	return nil
}

// NewMonitorFilter produces a Falcore RequestFilter for a given named service.
// This filter will forward to the service if it is up, otherwise it will
// display an error page to the requester. There are also optional triggers
// which would be run if the service is down (presumably to bring it up), or if
// the service is already up, or in either case, respectively.
func (monitor *MinMonitor) NewMinMonitorFilter(
	name string,
) (falcore.RequestFilter, error) {
	svc, serviceExists := monitor.table[name]
	if ! serviceExists {
		log().Err(
			fmt.Sprintf(
				"minmonitor cannot create a request filter" +
				" for unknown service: \"%s\"",
				name,
			),
		)

		return nil, UnknownServiceError
	}

	return svc, nil
}

func (svc *MinMonitorredService) FilterRequest(
	req *falcore.Request,
) (*http.Response) {
	log().Debug("running minmonitor filter")

	up, err := svc.Status()
	if err != nil {
		log().Warning(
			fmt.Sprintf(
				"minmonitor filter received an error" +
				" while requesting the status for \"%s:%d\":" +
				" %v",
				svc.Address,
				svc.Port,
				err,
			),
		)
		return falcore.StringResponse(
			req.HttpRequest,
			500,
			nil,
			"<html><head><title>Pullcord - Internal" +
			" Server Error</title></head><body><h1>" +
			"Pullcord - Internal Server Error</h1><p>An" +
			" internal server error has occurred, but it" +
			" might not be serious. However, If the" +
			" problem persists, the site administrator" +
			" should be contacted.</p></body></html>",
		)
	}

	if svc.Always != nil {
		err = svc.Always.Trigger()
		if err != nil {
			log().Warning(
				fmt.Sprintf(
					"minmonitor filter received" +
					" an error while running the" +
					" always trigger on \"%s:%d\": %v",
					svc.Address,
					svc.Port,
					err,
				),
			)
			return falcore.StringResponse(
				req.HttpRequest,
				500,
				nil,
				"<html><head><title>Pullcord -" +
				" Internal Server Error</title>" +
				"</head><body><h1>Pullcord -" +
				" Internal Server Error</h1><p>" +
				"An internal server error has" +
				" occurred, but it might not be" +
				" serious. However, if the problem" +
				" persists, the site administrator" +
				" should be contacted.</p></body>" +
				"</html>",
			)
		}
	}

	if up {
		if svc.OnUp != nil {
			err = svc.OnUp.Trigger()
			if err != nil {
				log().Warning(
					fmt.Sprintf(
						"minmonitor filter" +
						" received an error" +
						" while running the" +
						" onDown trigger on" +
						" \"%s:%d\": %v",
						svc.Address,
						svc.Port,
						err,
					),
				)
				return falcore.StringResponse(
					req.HttpRequest,
					500,
					nil,
					"<html><head><title>Pullcord" +
					" - Internal Server Error" +
					"</title></head><body><h1>" +
					"Pullcord - Internal Server" +
					" Error</h1><p>An internal" +
					" server error has occurred," +
					" but it might not be" +
					" serious. However, if the" +
					" problem persists, the site" +
					" administrator should be" +
					" contacted.</p></body></html>",
				)
			}
		}

		log().Debug("minmonitor filter passthru")
		return svc.passthru.FilterRequest(req)
	}

	if svc.OnDown != nil {
		err = svc.OnDown.Trigger()
		if err != nil {
			log().Warning(
				fmt.Sprintf(
					"minmonitor filter received" +
					" an error while running the" +
					" onDown trigger on \"%s:%d\": %v",
					svc.Address,
					svc.Port,
					err,
				),
			)
			return falcore.StringResponse(
				req.HttpRequest,
				500,
				nil,
				"<html><head><title>Pullcord -" +
				" Internal Server Error</title>" +
				"</head><body><h1>Pullcord -" +
				" Internal Server Error</h1><p>" +
				"An internal server error has" +
				" occurred, but it might not be" +
				" serious. However, If the problem" +
				" persists, the site administrator" +
				" should be contacted.</p></body>" +
				"</html>",
			)
		}
	}

	log().Info(
		fmt.Sprintf(
			"minmonitor filter has reached a down" +
			" service (\"%s:%d\"), but any triggers have" +
			" fired successfully",
			svc.Address,
			svc.Port,
		),
	)
	return falcore.StringResponse(
		req.HttpRequest,
		503,
		nil,
		"<html><head><title>Pullcord - Service Not Ready"+
		"</title></head><body><h1>Pullcord - Service Not" +
		" Ready</h1><p>The requested service is not yet" +
		" ready, but any trigger to start the service has" +
		" been started successfully, so hopefully the" +
		" service will be up in a few minutes.</p><p>If you" +
		" would like further information, please contact the" +
		" site administrator.</p></body></html>",
	)
}

// NewMinMonitor constructs a new MinMonitor.
func NewMinMonitor() *MinMonitor {
	log().Info("initializing minimal service monitor")

	var result MinMonitor
	result.table = make(map[string]*MinMonitorredService)

	return &result
}
