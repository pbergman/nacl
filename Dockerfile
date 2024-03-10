FROM ubuntu:16.04

RUN dpkg --add-architecture mipsel
RUN echo "deb [trusted=yes] http://ftp.de.debian.org/debian buster main" > /etc/apt/sources.list.d/buster.list
RUN apt update
RUN apt install -y \
    build-essential \
    gcc-mipsel-linux-gnu \
    libpcap-dev:mipsel \
    libpcap0.8-dev:mipsel \
    libc6-dev:mipsel \
    libdbus-1-dev:mipsel \
    libpcap0.8:mipsel \
    wget
RUN wget -c https://go.dev/dl/go1.19.8.linux-amd64.tar.gz -O - | tar -xz -C  /usr/local

ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH