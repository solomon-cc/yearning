FROM golang:1.16 as build

# 容器环境变量添加，会覆盖默认的变量值
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作区
WORKDIR /go/release
ADD .  .

RUN go build -o yearning


FROM alpine:3.12

LABEL maintainer="Solomon.-2020/11/16"

EXPOSE 8000

COPY --from=build /go/release/yearning  /opt/yearning
COPY conf.toml /opt/conf.toml
ENV DEV="true"

RUN echo "http://mirrors.ustc.edu.cn/alpine/v3.12/main/" > /etc/apk/repositories && \
      apk add --no-cache tzdata libc6-compat && \
      ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
      echo "Asia/Shanghai" >> /etc/timezone && \
      echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf

WORKDIR /opt

CMD /opt/yearning install && /opt/yearning run -b "https://sqlweb.ayla.com.cn"
