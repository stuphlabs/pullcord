package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/authentication"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/monitor"
	pcnet "github.com/stuphlabs/pullcord/net"
	"github.com/stuphlabs/pullcord/proxy"
	"github.com/stuphlabs/pullcord/trigger"
	"github.com/stuphlabs/pullcord/util"
)

const defaultConfigFilePath = "/etc/pullcord.json"
const defaultConfig = `{
	"resources": {
		"hresource": {
			"type": "exactpathrouter",
			"data": {
				"routes": {
					"/favicon.ico": {
						"type": "standardresponse",
						"data": 404
					}
				},
				"default": {
					"type": "landingfaulter",
					"data": {}
				}
			}
		},
		"lresource": {
			"type": "basiclistener",
			"data": {
				"proto": "tcp",
				"laddr": ":8080"
			}
		}
	},
	"listener": "lresource",
	"handler": "hresource"
}`

func main() {
	var inlineCfg string
	var cfgPath string
	var cfgFallback bool

	flag.StringVar(
		&inlineCfg,
		"inline-config",
		"",
		"Inline pullcord config instead of using a config file",
	)

	flag.StringVar(
		&cfgPath,
		"config",
		defaultConfigFilePath,
		"Path to pullcord config file",
	)

	flag.BoolVar(
		&cfgFallback,
		"config-fallback",
		true,
		"Fallback to basic config if unable to find the config file",
	)

	flag.Parse()

	var err error
	var cfgReader io.Reader
	if inlineCfg != "" {
		cfgReader = strings.NewReader(inlineCfg)
	}

	if cfgReader == nil {
		cfgReader, err = os.Open(cfgPath)
		if err != nil {
			log.Error(
				fmt.Sprintf(
					"Unable to open specified config file"+
						" %s: %s",
					cfgPath,
					err.Error(),
				),
			)
		} else {
			log.Info(
				fmt.Sprintf(
					"Reading config from file: %s",
					cfgPath,
				),
			)
		}
	}

	if cfgReader == nil {
		if !cfgFallback {
			log.Crit(
				"No config defined and not falling back to"+
					" default, aborting.",
			)
			os.Exit(1)
		} else {
			cfgReader = strings.NewReader(defaultConfig)
		}
	}

	authentication.LoadPlugin()
	monitor.LoadPlugin()
	pcnet.LoadPlugin()
	proxy.LoadPlugin()
	trigger.LoadPlugin()
	util.LoadPlugin()

	log.Debug("Plugins loaded")

	server, err := config.ServerFromReader(cfgReader)
	if err != nil {
		log.Debug(err)
		critErr := fmt.Errorf(
			"Error while parsing server config: %s",
			err.Error(),
		)
		log.Crit(critErr)
		panic(critErr)
	}

	log.Notice(
		fmt.Sprintf(
			"Starting server at %s...",
			server.Listener.Addr(),
		),
	)

	err = server.Serve()
	if err != nil {
		log.Debug(err)
		critErr := fmt.Errorf(
			"Error while running server: %s",
			err.Error(),
		)
		log.Crit(critErr)
		panic(critErr)
	}
}
