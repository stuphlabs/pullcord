package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/authentication"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/monitor"
	pcnet "github.com/stuphlabs/pullcord/net"
	"github.com/stuphlabs/pullcord/proxy"
	"github.com/stuphlabs/pullcord/trigger"
	"github.com/stuphlabs/pullcord/util"
)

const DefaultConfigFilePath = "/etc/pullcord.json"

func main() {
	cfgPath := flag.String(
		"config",
		DefaultConfigFilePath,
		"Path to pullcord config file",
	)

	cfgFallback := flag.Bool(
		"config-fallback",
		false,
		"Fallback to basic config if unable to find the config file",
	)

	flag.Parse()

	if cfgPath == nil {
		log.Warning(
			"Command line flag for the config file did not parse" +
				" as expected.",
		)
		cfgPathVal := DefaultConfigFilePath
		cfgPath = &cfgPathVal
	}

	if cfgFallback == nil {
		log.Warning(
			"Command line flag for the config fallback did not" +
				" parse as expected.",
		)
		cfgFallbackVal := true
		cfgFallback = &cfgFallbackVal
	}

	authentication.LoadPlugin()
	monitor.LoadPlugin()
	pcnet.LoadPlugin()
	proxy.LoadPlugin()
	trigger.LoadPlugin()
	util.LoadPlugin()

	var server *config.Server
	cfgReader, err := os.Open(*cfgPath)
	if err != nil {
		log.Debug(err)

		if !*cfgFallback {
			critErr := fmt.Errorf(
				"Unable to open config file: %s\nConsider"+
					" using --config-fallback.",
				err.Error(),
			)
			log.Crit(critErr)
			panic(critErr)
		}

		log.Notice(
			fmt.Sprintf(
				"Falling back to a basic config since config"+
					" file could not be opened: %s",
				err.Error(),
			),
		)

		var nl net.Listener
		nl, err = net.Listen("tcp", ":80")
		if err != nil {
			log.Debug(err)
			critErr := fmt.Errorf(
				"Unable to open port 80 for the fallback"+
					" config: %s",
				err.Error(),
			)
			log.Crit(critErr)
			panic(critErr)
		}

		handler := &config.ConfigPipeline{
			Server: falcore.NewServer(0, falcore.NewPipeline()),
		}
		landingFilter := falcore.RequestFilter(new(util.LandingFilter))
		handler.Server.Pipeline.Upstream.PushBack(
			&util.ExactPathRouter{
				Routes: map[string]*falcore.RequestFilter{
					"/": &landingFilter,
				},
			},
		)
		server = &config.Server{
			Handler: handler,
			Listener: &pcnet.BasicListener{
				Listener: nl,
			},
		}
	} else {
		log.Info(fmt.Sprintf("Reading config from file: %s", *cfgPath))
		server, err = config.ServerFromReader(cfgReader)
		if err != nil {
			log.Debug(err)
			critErr := fmt.Errorf(
				"Error while parsing server config: %s",
				err.Error(),
			)
			log.Crit(critErr)
			panic(critErr)
		}
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
