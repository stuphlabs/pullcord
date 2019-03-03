package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/proidiot/gone/errors"
	"github.com/proidiot/gone/log"

	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/proxy"
	"github.com/stuphlabs/pullcord/trigger"
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

// MinMonitorredService holds the information for a single service definition.
type MinMonitorredService struct {
	URL         *url.URL
	GracePeriod time.Duration
	OnDown      trigger.Triggerrer
	OnUp        trigger.Triggerrer
	Always      trigger.Triggerrer
	lastChecked time.Time
	up          bool
	passthru    http.Handler
}

func init() {
	config.MustRegisterResourceType(
		"minmonitorredservice",
		func() json.Unmarshaler {
			return new(MinMonitorredService)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (s *MinMonitorredService) UnmarshalJSON(data []byte) error {
	var t struct {
		URL         string
		GracePeriod string
		OnDown      *config.Resource
		OnUp        *config.Resource
		Always      *config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	g, e := time.ParseDuration(t.GracePeriod)
	if e != nil {
		return e
	}

	s.GracePeriod = g

	if t.OnDown != nil {
		d := t.OnDown.Unmarshalled
		switch d := d.(type) {
		case trigger.Triggerrer:
			s.OnDown = d
		default:
			return config.UnexpectedResourceType
		}
	} else {
		// TODO test null values for these as well
		s.OnDown = nil
	}

	if t.OnUp != nil {
		u := t.OnUp.Unmarshalled
		switch u := u.(type) {
		case trigger.Triggerrer:
			s.OnUp = u
		default:
			return config.UnexpectedResourceType
		}
	} else {
		s.OnUp = nil
	}

	if t.Always != nil {
		a := t.Always.Unmarshalled
		switch a := a.(type) {
		case trigger.Triggerrer:
			s.Always = a
		default:
			return config.UnexpectedResourceType
		}
	} else {
		s.Always = nil
	}

	u, e := url.Parse(t.URL)
	if e != nil {
		return e
	}

	s.URL = u

	return nil
}

// NewMinMonitorredService creates an initialized MinMonitorredService.
func NewMinMonitorredService(
	u *url.URL,
	gracePeriod time.Duration,
	onDown trigger.Triggerrer,
	onUp trigger.Triggerrer,
	always trigger.Triggerrer,
) (service *MinMonitorredService, err error) {
	result := MinMonitorredService{
		URL:         u,
		GracePeriod: gracePeriod,
		OnDown:      onDown,
		OnUp:        onUp,
		Always:      always,
		lastChecked: time.Time{},
		up:          false,
		passthru:    nil,
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
		_ = log.Err(
			fmt.Sprintf(
				"minmonitor cannot add service: name \"%s\""+
					" previously used for service at"+
					" \"%s\", new service at \"%s\"",
				name,
				osvc.URL.String(),
				service.URL.String(),
			),
		)

		return DuplicateServiceRegistrationError
	}

	if monitor.table == nil {
		monitor.table = make(map[string]*MinMonitorredService)
	}

	monitor.table[name] = service

	_ = log.Info(
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
	s, entryExists := monitor.table[name]
	if !entryExists {
		_ = log.Err(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:"+
					" \"%s\"",
				name,
			),
		)

		return false, UnknownServiceError
	}

	return s.Reprobe()
}

// Reprobe forces the status of the service to be checked immediately without
// regard to a possible previously cached up status. The result of this probe
// will automatically be cached by the monitor.
func (s *MinMonitorredService) Reprobe() (up bool, err error) {
	hostname := s.URL.Hostname()
	socktypefam := "tcp"
	if strings.Index(hostname, ":") > 0 {
		hostname = fmt.Sprintf("//[%s]", hostname)
		socktypefam = "tcp6"
	}
	port := s.URL.Port()
	if port == "" {
		port = s.URL.Scheme
	}

	addr := fmt.Sprintf("%s:%s", hostname, port)
	conn, err := net.Dial(socktypefam, addr)
	s.lastChecked = time.Now()
	if err != nil {
		s.up = false
		// TODO check what the error was

		switch castErr := err.(type) {
		case *net.OpError:
			if castErr.Addr != nil {
				_ = log.Info(
					fmt.Sprintf(
						"minmonitor received a"+
							" connection refused"+
							" (interpereted as a"+
							" down status) from"+
							" \"%s\"",
						s.URL.String(),
					),
				)

				return false, nil
			}

			_ = log.Warning(
				fmt.Sprintf(
					"minmonitor encountered an error while"+
						" probing \"%s\": %v",
					s.URL.String(),
					err,
				),
			)

			return false, err
		default:
			_ = log.Warning(
				fmt.Sprintf(
					"minmonitor encountered an unknown"+
						" error while probing"+
						" \"%s\": %v",
					s.URL.String(),
					err,
				),
			)

			return false, err
		}
	} else {
		defer func() {
			_ = conn.Close()
		}()
		s.up = true

		_ = log.Info(
			fmt.Sprintf(
				"minmonitor successfully probed: \"%s\"",
				s.URL.String(),
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
	s, entryExists := monitor.table[name]
	if !entryExists {
		_ = log.Err(
			fmt.Sprintf(
				"minmonitor cannot probe unknown service:"+
					" \"%s\"",
				name,
			),
		)

		return false, UnknownServiceError
	}

	return s.Status()
}

// Status returns true if the status of the service is currently believed to be
// up. The service could have its status reported as being up if a brand new
// probe of the service indicates that the service is indeed up, or if a recent
// probe indicated that the service was up (specifically if the most recent
// probe indicated that the service was up and that probe was within a grace
// period that was specified when the service was registered), or if the status
// of the service was explicitly set as being up within that same grace period
// (and no other forced re-probe has occurred since this forced status up
// assignment). However, if the status of the service is reported as being down,
// then it necessarily means that a probe has just occurred and the service was
// unable to be reached.
func (s *MinMonitorredService) Status() (up bool, err error) {
	if (!s.up) || time.Now().After(
		s.lastChecked.Add(s.GracePeriod),
	) {
		_ = log.Info(
			fmt.Sprintf(
				"minmonitor must reprobe as either the grace"+
					" period has lapsed or the previous"+
					" probe indicated a down status for:"+
					" \"%s\"",
				s.URL.String(),
			),
		)

		return s.Reprobe()
	}

	_ = log.Info(
		fmt.Sprintf(
			"minmonitor is skipping the reprobe as the current"+
				" time is still within the grace period of the"+
				" last successfull probe of: \"%s\"",
			s.URL.String(),
		),
	)

	return true, nil
}

// SetStatusUp explicitly sets the status of a named service as being up. This
// up status will be cached just as if it were the result of a normal probe.
func (monitor *MinMonitor) SetStatusUp(name string) (err error) {
	s, entryExists := monitor.table[name]
	if !entryExists {
		_ = log.Err(
			fmt.Sprintf(
				"minmonitor cannot set the status of unknown"+
					" service: \"%s\"",
				name,
			),
		)

		return UnknownServiceError
	}

	return s.SetStatusUp()
}

// SetStatusUp explicitly sets the status of the service as being up. This up
// status will be cached just as if it were the result of a normal probe.
func (s *MinMonitorredService) SetStatusUp() error {
	_ = log.Info(
		fmt.Sprintf(
			"minmonitor has been explicitly informed of the up"+
				" status of: \"%s\"",
			s.URL.String(),
		),
	)
	s.lastChecked = time.Now()
	s.up = true

	return nil
}

// NewMinMonitorFilter produces an http.Handler for a given named service. This
// handler will forward to the service if it is up, otherwise it will display an
// error page to the requester. There are also optional triggers which would be
// run if the service is down (presumably to bring it up), or if the service is
// already up, or in either case, respectively.
func (monitor *MinMonitor) NewMinMonitorFilter(
	name string,
) (http.Handler, error) {
	s, serviceExists := monitor.table[name]
	if !serviceExists {
		_ = log.Err(
			fmt.Sprintf(
				"minmonitor cannot create a request filter"+
					" for unknown service: \"%s\"",
				name,
			),
		)

		return nil, UnknownServiceError
	}

	return s, nil
}

func (s *MinMonitorredService) ServeHTTP(
	w http.ResponseWriter,
	req *http.Request,
) {
	_ = log.Debug("running minmonitor filter")

	up, err := s.Status()
	if err != nil {
		_ = log.Warning(
			fmt.Sprintf(
				"minmonitor filter received an error"+
					" while requesting the status for"+
					" \"%s\": %v",
				s.URL.String(),
				err,
			),
		)
		w.WriteHeader(500)
		_, err = fmt.Fprint(
			w,
			"<html><head><title>Pullcord - Internal"+
				" Server Error</title></head><body><h1>"+
				"Pullcord - Internal Server Error</h1><p>An"+
				" internal server error has occurred, but it"+
				" might not be serious. However, If the"+
				" problem persists, the site administrator"+
				" should be contacted.</p></body></html>",
		)
		if err != nil {
			_ = log.Error(
				fmt.Sprintf(
					"error writing page after error with"+
						" monitor status: %s",
					err.Error(),
				),
			)
		}
		return
	}

	if s.Always != nil {
		_ = log.Debug("minmonitor running always trigger")
		err = s.Always.Trigger()
		if err != nil {
			_ = log.Warning(
				fmt.Sprintf(
					"minmonitor filter received"+
						" an error while running the"+
						" always trigger on \"%s\":"+
						" %v",
					s.URL.String(),
					err,
				),
			)
			w.WriteHeader(500)
			_, err = fmt.Fprint(
				w,
				"<html><head><title>Pullcord -"+
					" Internal Server Error</title>"+
					"</head><body><h1>Pullcord -"+
					" Internal Server Error</h1><p>"+
					"An internal server error has"+
					" occurred, but it might not be"+
					" serious. However, if the problem"+
					" persists, the site administrator"+
					" should be contacted.</p></body>"+
					"</html>",
			)
			if err != nil {
				_ = log.Error(
					fmt.Sprintf(
						"error writing page after"+
							" error with always"+
							" trigger: %s",
						err.Error(),
					),
				)
			}
			return
		}
		_ = log.Debug("minmonitor completed always trigger")
	}

	if up {
		_ = log.Debug("minmonitor determined service is up")
		if s.OnUp != nil {
			_ = log.Debug("minmonitor running up trigger")
			err = s.OnUp.Trigger()
			if err != nil {
				_ = log.Warning(
					fmt.Sprintf(
						"minmonitor filter"+
							" received an error"+
							" while running the"+
							" onDown trigger on"+
							" \"%s\": %v",
						s.URL.String(),
						err,
					),
				)
				w.WriteHeader(500)
				_, err = fmt.Fprint(
					w,
					"<html><head><title>Pullcord"+
						" - Internal Server Error"+
						"</title></head><body><h1>"+
						"Pullcord - Internal Server"+
						" Error</h1><p>An internal"+
						" server error has occurred,"+
						" but it might not be"+
						" serious. However, if the"+
						" problem persists, the site"+
						" administrator should be"+
						" contacted.</p></body></html>",
				)
				if err != nil {
					_ = log.Error(
						fmt.Sprintf(
							"error writing page"+
								" after error"+
								" with onup"+
								" trigger: %s",
							err.Error(),
						),
					)
				}
				return
			}
			_ = log.Debug("minmonitor completed up trigger")
		}

		if s.passthru == nil {
			_ = log.Debug(
				"minmonitor filter passthru creation started",
			)
			s.passthru = proxy.NewPassthruFilter(s.URL)
			_ = log.Debug(
				"minmonitor filter passthru creation completed",
			)
		}

		_ = log.Debug("minmonitor filter passthru starting")
		s.passthru.ServeHTTP(w, req)
		_ = log.Debug("minmonitor filter passthru completed")
		return
	}

	_ = log.Debug("minmonitor determined service is down")
	if s.OnDown != nil {
		_ = log.Debug("minmonitor running down trigger")
		err = s.OnDown.Trigger()
		if err != nil {
			_ = log.Warning(
				fmt.Sprintf(
					"minmonitor filter received"+
						" an error while running the"+
						" onDown trigger on \"%s\": %v",
					s.URL.String(),
					err,
				),
			)
			w.WriteHeader(500)
			_, err = fmt.Fprint(
				w,
				"<html><head><title>Pullcord -"+
					" Internal Server Error</title>"+
					"</head><body><h1>Pullcord -"+
					" Internal Server Error</h1><p>"+
					"An internal server error has"+
					" occurred, but it might not be"+
					" serious. However, If the problem"+
					" persists, the site administrator"+
					" should be contacted.</p></body>"+
					"</html>",
			)
			if err != nil {
				_ = log.Error(
					fmt.Sprintf(
						"error writing page after"+
							" error with ondown"+
							" trigger: %s",
						err.Error(),
					),
				)
			}
			return
		}
		_ = log.Debug("minmonitor completed down trigger")
	}

	_ = log.Info(
		fmt.Sprintf(
			"minmonitor filter has reached a down"+
				" service (\"%s\"), but any triggers have"+
				" fired successfully",
			s.URL.String(),
		),
	)
	w.WriteHeader(503)
	_, err = fmt.Fprint(
		w,
		"<html><head><title>Pullcord - Service Not Ready"+
			"</title></head><body><h1>Pullcord - Service Not"+
			" Ready</h1><p>The requested service is not yet"+
			" ready, but any trigger to start the service has"+
			" been started successfully, so hopefully the"+
			" service will be up in a few minutes.</p><p>If you"+
			" would like further information, please contact the"+
			" site administrator.</p></body></html>",
	)
	if err != nil {
		_ = log.Error(
			fmt.Sprintf(
				"error writing page after status down: %s",
				err.Error(),
			),
		)
	}
	return
}

// NewMinMonitor constructs a new MinMonitor.
func NewMinMonitor() *MinMonitor {
	_ = log.Info("initializing minimal service monitor")

	var result MinMonitor
	result.table = make(map[string]*MinMonitorredService)

	return &result
}
