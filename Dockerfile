FROM golang:1.22.4-bullseye AS builder

# ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
# ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
# ENV CGO_ENABLED=1
# ENV GO111MODULE=on

WORKDIR /workspace
COPY . .
RUN  make build


FROM golang:1.22.4-bullseye
RUN apt-get update && apt-get install -y jq mariadb-client
COPY --from=builder /workspace/build/mechain-sp /usr/bin/mechain-sp
CMD ["mechain-sp"]