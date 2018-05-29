# redisSky
FROM node:6

RUN apt-get update \
    && apt-get install -y wget;

RUN mkdir /app

WORKDIR /app

RUN wget https://www.golangtc.com/static/go/1.9.2/go1.9.2.linux-amd64.tar.gz \
    && tar zxvf go1.9.2.linux-amd64.tar.gz -C /usr/local \
    && mkdir -p /app/go/src \
    && mkdir /app/go/pkg \
    && mkdir /app/go/bin;

ENV GOPATH=/app/go
ENV GOROOT=/usr/local/go
ENV GOBIN=/usr/local/go/bin/
ENV PATH=$PATH:$GOBIN:/app/go/bin


RUN mkdir -p /app/go/src/github.com/prettyyjnic

WORKDIR /app/go/src/github.com/prettyyjnic
RUN git clone https://gitee.com/stuinfer/redisSky.git

RUN cd /app/go/src/github.com/prettyyjnic/redisSky/frontend \
    && npm install \
    && npm run build;

RUN cd /app/go/src/github.com/prettyyjnic/redisSky/backend/bin \
    && go build start.go;

WORKDIR /app/go/src/github.com/prettyyjnic/redisSky/backend/bin
CMD ["./start"]