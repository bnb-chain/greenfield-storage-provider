# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build storage provider in a stock Go builder container
FROM golang:1.19-alpine as builder

RUN apk add --no-cache make git bash

ADD . /storageprovider

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN cd /storageprovider && make build

# Pull storage provider into a second stage deploy alpine container
FROM alpine:3.16.0

ARG SP_USER=sp
ARG SP_USER_UID=1000
ARG SP_USER_GID=1000

ENV PACKAGES ca-certificates bash curl libstdc++
ENV WORKDIR=/server

RUN apk add --no-cache $PACKAGES \
  && rm -rf /var/cache/apk/* \
  && addgroup -g ${SP_USER_GID} ${SP_USER} \
  && adduser -u ${SP_USER_UID} -G ${SP_USER} --shell /sbin/nologin --no-create-home -D ${SP_USER} \
  && addgroup ${SP_USER} tty \
  && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd

RUN echo "[ ! -z \"\$TERM\" -a -r /etc/motd ] && cat /etc/motd" >> /etc/bash/bashrc

WORKDIR ${WORKDIR}

COPY --from=builder /storageprovider/build/bin/storage_provider ${WORKDIR}/
RUN chown -R ${SP_USER_UID}:${SP_USER_GID} ${WORKDIR}
USER ${SP_USER_UID}:${SP_USER_GID}

EXPOSE 9033 9133 9233 9333 9433 9533

ENTRYPOINT ["/server/storage_provider"]
