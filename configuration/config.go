package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	"github.com/stuphlabs/pullcord"
	"github.com/stuphlabs/pullcord/authentication"
	"github.com/stuphlabs/pullcord/registry"
	"github.com/stuphlabs/pullcord/monitor"
	"github.com/stuphlabs/pullcord/proxy"
	"github.com/stuphlabs/pullcord/trigger"
	"github.com/stuphlabs/pullcord/util"
	"runtime/debug"
)

type ConfigResource struct {
	Name string
	Type string
	Data json.RawMessage
}

type Config struct {
	Resources []*ConfigResource
	Pipeline []string
	Port int
}

func ServerFromString(jsonData []byte) (*falcore.Server, error) {
	var config Config
	if err := json.Unmarshal(jsonData, &config); err != nil {
		log().Crit(
			fmt.Sprintf(
				"Unable to parse config file: %v",
				err,
			),
		)
		return nil, err
	}

	if e := registry.Init(); e != nil {
		return nil, e
	}

	for _, r := range config.Resources {
		var dst interface{}
		switch r.Type {
		case "landingfilter":
			dst = new(pullcord.LandingFilter)
		case "cookiemaskfilter":
			dst = new(authentication.CookiemaskFilter)
		case "inmempwdstore":
			dst = new(authentication.InMemPwdStore)
		case "loginhandler":
			dst = new(authentication.LoginHandler)
		case "minsessionhandler":
			dst = new(authentication.MinSessionHandler)
		case "minmonitor":
			dst = new(monitor.MinMonitor)
		case "passthrufilter":
			dst = new(proxy.PassthruFilter)
		case "compundtrigger":
			dst = new(trigger.CompoundTrigger)
		case "delaytrigger":
			dst = new(trigger.DelayTrigger)
		case "ratelimittrigger":
			dst = new(trigger.RateLimitTrigger)
		case "shelltrigger":
			dst = new(trigger.ShellTriggerHandler)
		case "standardresponse":
			dst = new(util.StandardResponse)
		case "exactpathrouter":
			dst = new(util.ExactPathRouter)
		default:
			log().Crit(
				fmt.Sprintf(
					"Unknown resource type: %s",
					r.Type,
				),
			)
			// TODO remove
			debug.PrintStack()
			return nil, errors.New(
				fmt.Sprintf(
					"Unknown resource type: %s",
					r.Type,
				),
			)
		}
		if err := json.Unmarshal(r.Data, dst); err != nil {
			log().Crit(
				fmt.Sprintf(
					"Unable to parse data for resource:" +
					" %s: %v",
					r.Name,
					err,
				),
			)
			return nil, err
		}
		log().Debug(
			fmt.Sprintf(
				"Successfully registered resource: %s",
				r.Name,
			),
		)
		if e := registry.Set(r.Name, dst); e != nil {
			return nil, e
		}
	}

	pipeline := falcore.NewPipeline()

	for _, upstream := range config.Pipeline {
		r, e := registry.Get(upstream)
		if e != nil {
			log().Crit(
				fmt.Sprintf(
					"Requested pipeline resource not" +
					" found: %s",
					upstream,
				),
			)
			return nil, e
		}
		switch r := r.(type) {
		case falcore.Router:
		case falcore.RequestFilter:
			pipeline.Upstream.PushBack(r)
		default:
			log().Crit(
				fmt.Sprintf(
					"Requested pipeline resource not a" +
					" RequestFilter: %s",
					upstream,
				),
			)
			return nil, registry.UnexpectedType
		}
		log().Debug(
			fmt.Sprintf(
				"Added upstream resource: %s",
				upstream,
			),
		)
	}

	server := falcore.NewServer(config.Port, pipeline)

	registry.Wipe()

	return server, nil
}

