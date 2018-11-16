FROM ubuntu:latest

RUN apt-get update -y && apt-get install -y \
	ca-certificates

COPY bin/pullcord /usr/bin/pullcord
COPY example /example
VOLUME /etc

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/usr/bin/pullcord"]
CMD ["--config","/example/basic.json"]
