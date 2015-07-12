[![Stories in Ready](https://badge.waffle.io/stuphlabs/pullcord.png?label=ready&title=Ready)](https://waffle.io/stuphlabs/pullcord)
Scale Down
==========

[![Build Status](https://drone.io/github.com/stuphlabs/pullcord/status.png)](https://drone.io/github.com/stuphlabs/pullcord/latest)
[![Build Status](https://travis-ci.org/stuphlabs/pullcord.svg)](https://travis-ci.org/stuphlabs/pullcord)

**Scale Down** will be a reverse proxy for cloud-based web apps that will allow the servers the web apps run on to be turned off when not in use. Scale Down should be nimble enough to run on the smallest cloud servers available, but it will be able to quickly spin up much larger servers running much bulkier web apps. Once traffic to these larger servers has stopped for a specified period of time, these other servers will be spun back down. Scale Down will be able to perform these duties for multiple web apps simultaneously, so there is no need to have more than one Scale Down server, and there is no need to consolidate multiple web apps onto a single server (in fact, Scale Down will likely work better if each web app is on its own server, as it will still be possible to spin up multiple servers if there are dependencies between the web apps).

For more information, email [Charlie](mailto://charlie@stuphlabs.com) or [Michael](mailto://michael@stuphlabs.com).
