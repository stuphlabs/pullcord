Pullcord
========

[![Build Status](https://img.shields.io/travis/stuphlabs/pullcord/master.svg)](https://travis-ci.org/stuphlabs/pullcord)
[![Coverage](https://img.shields.io/coveralls/stuphlabs/pullcord/master.svg)](https://coveralls.io/github/stuphlabs/pullcord?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/stuphlabs/pullcord)](https://goreportcard.com/report/github.com/stuphlabs/pullcord)
[![Code Climate](https://img.shields.io/codeclimate/github/stuphlabs/pullcord.svg)](https://codeclimate.com/github/stuphlabs/pullcord)
[![Open Issues](https://img.shields.io/github/issues-raw/stuphlabs/pullcord.svg)](https://waffle.io/stuphlabs/pullcord)
[![Godoc](http://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.org/github.com/stuphlabs/pullcord)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/982/badge)](https://bestpractices.coreinfrastructure.org/projects/982)


**Pullcord** will be a reverse proxy for cloud-based web apps that will allow
the servers the web apps run on to be turned off when not in use. Pullcord
should be nimble enough to run on the smallest cloud servers available, but it
will be able to quickly spin up much larger servers running much bulkier web
apps. Once traffic to these larger servers has stopped for a specified period
of time, these other servers will be spun back down. Pullcord will be able to
perform these duties for multiple web apps simultaneously, so there is no need
to have more than one Pullcord server, and there is no need to consolidate
multiple web apps onto a single server (in fact, Pullcord will likely work
better if each web app is on its own server, as it will still be possible to
spin up multiple servers if there are dependencies between the web apps).

For more information, email [Charlie](mailto://charlie@stuphlabs.com).

## Acceptance Testing
To aide in acceptance testing, first run the tests and build the container from
a clean copy of the codebase:
```
make clean container
```

Then run the pullcord container with the default config:
```
docker run \
	-d \
	--name pullcord-acceptance \
	-p 127.0.0.1:8080:8080 \
	-e LOG_UPTO="LOG_DEBUG" \
	pullcord
```

Now visit [your local pullcord instance](http://127.0.0.1:8080/).

To follow along with the logs:
```
docker logs -f pullcord-acceptance
```

To clean up and collect logs when finished:
```
docker kill pullcord-acceptance || echo "Container already killed, continuing."

docker logs pullcord-acceptance > pullcord-acceptance-`date +%s`.log

docker rm pullcord-acceptance
```

To try a different config, you could set a different `${PULLCORD_CONFIG_PATH}`
in this command (be sure it evaluates to a full path):
```
PULLCORD_CONFIG_PATH="${PWD}/example/basic.json"
docker run \
	-d \
	--name pullcord-acceptance \
	-p 127.0.0.1:8080:8080 \
	-e LOG_UPTO="LOG_DEBUG" \
	-v `dirname ${PULLCORD_CONFIG_PATH}`:/config \
	pullcord --config /config/`basename ${PULLCORD_CONFIG_PATH}`
```


## Common make targets
Just clean up any lingering out-of-date artifacts:
```
make clean
```

Just run tests:
```
make test
```

Just build binary (will run tests if they have not yet been run):
```
make
```
or:
```
make all
```

Just build container (will build binary if needed):
```
make container
```

## The Main Problem
Over the years, Stuph Labs has used various web apps and other software daemons
for our side projects (i.e. Gitolite, Trac, OpenVPN, SFTP, etc.), but this has
required a server to be running all the time despite the fact that we'd only
use this server for a few random hours a month. At least now that cloud
computing has become more popular, we are no longer restricted to choosing
between expensive dedicated servers or Seriously over-provisioned and inflexible
shared hosting. However, manually going into the various cloud consoles and
turning servers on and off is a hassle at best, and even when another potential
user is sufficiently trusted that they are given copies of the administrative
credentials (often in an insecure way to begin with), it is unrealistic to
think that such a user would log in to start and stop these servers as needed
when doing so requires that they use an unfamiliar and very complicated
interface that they realize is full of buttons that, if accidentally pressed,
could incur extreme costs in a very short span of time. As a result, just as
before modern cloud computing was an option, we have often resorted to either
eating the hefty cost of a properly equipped server that is only used 1% of the
time, or else using a seriously under-powered server that causes a great deal of
frustration to the users and is still only used 1% of the time.
One of the things that modern cloud computing has given us is the ability to
quickly, easily, and automatically scale from as little as one server up to
thousands in a short amount of time, and then almost as quickly scale the
number of servers back down again. While this has enabled regularly utilized
services to have a server footprint that more accurately matches their needs
(and thus save a tremendous amount of money without sacrificing availability or
performance), the same cannot be said of very lightly utilized services at this
time.

## The Secondary Problem
There are a variety of reasons we may install pieces of software at some point,
but there are also many reasons we may choose not to update these pieces of
software (perhaps we are trying to decide on which version of the software to
use elsewhere, perhaps we are trying to test the scope and ease of exploitation
of a known vulnerability, or perhaps we are just too busy/lazy to get around to
updating every single piece of software we aren't sold on yet anyway). While
most all of the software that we wouldn't want to update would not be used for
legitimate data we would care about, we are certainly aware that data leakage
is by no means the only thing one should worry about when it comes to
information security. As a result, we often either spend a frustrating amount
of time updating software we haven't decided if we care about at all, or else
we choose not to install the software in the first place for fear of running
into this very predicament. Today it is common to install such pieces of
software in virtual machines either manually or using some tool like Vagrant,
but setting up such services in a way that many people can test them over the
internet is tedious, error-prone, and depends on an always on system with
reliable internet access, at which point we are right back to the very reasons
we use external hosting in the first place. We could just have a beefy server
that is always on and running VM host software, but it is a waste of such a
machine if it is only occasionally used, and it wouldn't be able to scale if
you wanted to run very many of these services at once for even a short time.

## The (Possible) Solution
It should be possible for Pullcord to sit on one small server and launch other
much more powerful cloud servers which host the desired pieces of software.
Once the desired pieces of software have gone unused for a period of time, the
cloud servers will be automatically turned back off. While it may take a few
minutes for these pieces of software to first become available, they will not
feel sluggish at all once they are running. Furthermore, the total cost should
remain low since these cloud servers were only used for a short time. Also, if
Pullcord is used as a proxy for these servers, then the potentially vulnerable
pieces of software would only be exposed to properly authenticated and
authorized users of the Pullcord service.

## Initial Design Considerations
The proposed solution would have some tedious aspects (i.e. an HTTP proxying
mechanism, a cookie handler, etc.), but the internal complexity should be
relatively low, and so a minimalist design with modular functionality would
likely be the best choice. At this point (which is admittedly after some
initial development work), it would appear that this solution could be split
into a few largely distinct components: the remote services monitoring system,
the remote service launching/destructing (trigger) system, the user
authentication system, the proxying system, and the configuration system.

## Initial Development Considerations
Many programming languages would be acceptable for this project, but I chose to
use Go as it seemed well-suited for the task, and as such this seemed like a
good opportunity to try out the language on a larger project than the original
Go tutorial. There is a Go library called Falcore which could prove useful.
Given the likely low algorithmic complexity involved with solving this problem,
it should be possible to develop a solution using many minimal changes in an
iterative development technique. This also has the advantage of lending the
process to test-driven development techniques. However, it is important to
continuously update the documentation, something which many developers
(including myself) have often been bad at. By adding some tests to the
continuous integration process to check that both code coverage and the
documentation ratio doesn't drop below a certain threshold, it should be
possible to keep me honest.
