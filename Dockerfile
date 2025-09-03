ARG REPO_URL=docker.io
ARG IMAGE_BUILDER=golang:1.23.0-alpine3.20
ARG IMAGE_RUNNER=alpine:3.20

## --- Builder

FROM ${REPO_URL}/${IMAGE_BUILDER} AS builder
LABEL stage=builder
WORKDIR /build

ARG GOPROXY_URL
ENV GOPROXY=${GOPROXY_URL}
SHELL [ "/bin/sh", "-ec" ]

RUN apk add --no-cache make

COPY go.mod go.sum Makefile ./
COPY cmd/ cmd/
COPY internal/ internal/

ENV GOOS=linux
ENV GOARCH=amd64

RUN cp $(go env GOROOT)/misc/wasm/wasm_exec.js .

RUN make build
RUN make build/wasm

## --- Runner

FROM ${REPO_URL}/${IMAGE_RUNNER}

ARG UID=1337
ARG GID=1337
ENV APPLICATION_HOME=/home/app \
    TARGET_USER=app \
    TARGET_GROUP=app

RUN apk add --no-cache tzdata ; \
    getent group ${GID} || addgroup --gid ${GID} ${TARGET_GROUP} ; \
    adduser -u ${UID} -g ${TARGET_GROUP} -D -S -h ${APPLICATION_HOME} ${TARGET_USER}
WORKDIR ${APPLICATION_HOME}
USER ${TARGET_USER}

COPY --from=builder /tmp/bin/server /usr/bin/server
COPY --from=builder --chown=${TARGET_USER}:${TARGET_GROUP} /tmp/bin/game.wasm    ./
COPY --from=builder --chown=${TARGET_USER}:${TARGET_GROUP} /build/wasm_exec.js   ./
COPY --chown=${TARGET_USER}:${TARGET_GROUP} web/*.html ./

COPY --chown=${TARGET_USER}:${TARGET_GROUP} entrypoint.sh ./

CMD [ "/bin/sh", "/home/app/entrypoint.sh", "/usr/bin/server" ]
