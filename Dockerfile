FROM ubuntu:latest

LABEL "maintainer"="Oleksandr Shuienko <oleksandr.shuienko@gmail.com>"

RUN apt-get update && apt-get install -y ca-certificates

ADD goshort_linux_amd64 /goshort

EXPOSE 80/tcp

ENTRYPOINT ["/goshort", "-listen", "0.0.0.0:80"]