FROM golang:1.20-alpine as builder

RUN apk add --no-cache make git bash protoc

ADD . /greenfield-storage-provider

ENV CGO_ENABLED=1
ENV GO111MODULE=on

# For Private REPO
ARG GH_TOKEN=""
RUN go env -w GOPRIVATE="github.com/bnb-chain/*"
RUN git config --global url."https://${GH_TOKEN}@github.com".insteadOf "https://github.com"

RUN apk add --no-cache build-base libc-dev

RUN cd /greenfield-storage-provider \
    && make build

# Pull greenfield into a second stage deploy alpine container
FROM alpine:3.17

ARG USER=sp
ARG USER_UID=1000
ARG USER_GID=1000

ENV PACKAGES libstdc++ ca-certificates bash curl
ENV WORKDIR=/app

RUN apk add --no-cache $PACKAGES \
  && rm -rf /var/cache/apk/* \
  && addgroup -g ${USER_GID} ${USER} \
  && adduser -u ${USER_UID} -G ${USER} --shell /sbin/nologin --no-create-home -D ${USER} \
  && addgroup ${USER} tty \
  && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd

RUN echo "[ ! -z \"\$TERM\" -a -r /etc/motd ] && cat /etc/motd" >> /etc/bash/bashrc

WORKDIR ${WORKDIR}

COPY --from=builder /greenfield-storage-provider/build/* ${WORKDIR}/
RUN chown -R ${USER_UID}:${USER_GID} ${WORKDIR}
USER ${USER_UID}:${USER_GID}

ENTRYPOINT ["/app/gnfd-sp"]
