package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/proidiot/gone/errors"
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
		"handler": {
			"type": "exactpathrouter",
			"data": {
				"routes": {
					"/favicon.ico": {
						"type": "standardresponse",
						"data": 404
					}
				},
				"default": {
					"type": "landinghandler",
					"data": {}
				}
			}
		},
		"listener": {
			"type": "basiclistener",
			"data": {
				"proto": "tcp",
				"laddr": ":8080"
			}
		}
	},
	"server": {
		"type": "httpserver",
		"data": {
			"listener": {
				"type": "ref",
				"data": "listener"
			},
			"handler": {
				"type": "ref",
				"data": "handler"
			}
		}
	}
}
`

func main() {
	var inlineCfg string
	var cfgPath string
	var cfgFallback bool
	var cfgPrint bool

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

	flag.BoolVar(
		&cfgPrint,
		"print-config",
		false,
		"Write the entire config to the logs at debug level",
	)

	flag.Parse()

	var err error
	var cfgReader io.ReadSeeker
	if inlineCfg != "" {
		cfgReader = strings.NewReader(inlineCfg)
	}

	if cfgReader == nil {
		cfgReader, err = os.Open(cfgPath)
		if err != nil {
			err = log.Error(
				fmt.Sprintf(
					"Unable to open specified config file"+
						" %s: %s",
					cfgPath,
					err.Error(),
				),
			)
			if err != nil {
				panic(err)
			}
			cfgReader = nil
		} else {
			err = log.Info(
				fmt.Sprintf(
					"Reading config from file: %s",
					cfgPath,
				),
			)
			if err != nil {
				panic(err)
			}
		}
	}

	if cfgReader == nil {
		if !cfgFallback {
			err = errors.New(
				"No config defined and not falling back to" +
					" default, aborting.",
			)
			_ = log.Crit(err)
			panic(err)
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

	_ = log.Debug("Plugins loaded")

	if cfgPrint {
		b := new(bytes.Buffer)
		_, err = b.ReadFrom(cfgReader)
		if err != nil {
			panic(err)
		}

		_ = log.Debug(fmt.Sprintf("Config is: %s", b.String()))
		_, err2 := cfgReader.Seek(0, io.SeekStart)
		if err2 != nil {
			critErr := fmt.Errorf(
				"Error while rewinding config reader after"+
					" printing to debug logs: %s",
				err.Error(),
			)
			_ = log.Crit(critErr)
			panic(critErr)
		}
	}

	cfgParser := config.Parser{Reader: cfgReader}
	server, err := cfgParser.Server()
	if err != nil {
		_ = log.Debug(err)
		critErr := fmt.Errorf(
			"Error while parsing server config: %s",
			err.Error(),
		)
		_ = log.Crit(critErr)
		panic(critErr)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Failed", r)
		}
	}()

	err = log.Info("Starting server...")
	if err != nil {
		panic(err)
	}
	err = server.Serve()
	if err != nil {
		_ = log.Debug(err)
		critErr := fmt.Errorf(
			"Error while running server: %s",
			err.Error(),
		)
		_ = log.Crit(critErr)
		panic(critErr)
	}
	err = log.Info("Server stopped")
	if err != nil {
		panic(err)
	}
}
