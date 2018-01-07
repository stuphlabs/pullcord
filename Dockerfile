FROM ubuntu

COPY bin/pullcord /usr/bin/pullcord
COPY example /example
VOLUME /etc

EXPOSE 80
EXPOSE 443

CMD ["/usr/bin/pullcord","--config","/example/basic.json"]
