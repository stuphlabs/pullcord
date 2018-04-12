FROM ubuntu

COPY bin/pullcord /usr/bin/pullcord
COPY example /example
VOLUME /etc

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/usr/bin/pullcord"]
CMD ["--config","/example/basic.json"]
