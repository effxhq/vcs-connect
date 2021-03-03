ARG GO_VERSION=1.15
ARG ALPINE_VERSION=3.12
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /go/src/vcs-connect

RUN apk -U upgrade && apk add build-base git curl

ARG GITHUB_TOKEN
ARG GITHUB_USER

RUN git config --global url."https://$GITHUB_USER:$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

COPY go.mod .
COPY go.sum .
COPY Makefile .
RUN make deps

COPY . .

ARG VERSION=""
ARG GIT_SHA=""
ARG TIMESTAMP=""
RUN make build VERSION="${VERSION}" GIT_SHA="${GIT_SHA}" TIMESTAMP="${TIMESTAMP}"

FROM alpine:${ALPINE_VERSION}

COPY --from=builder /go/src/vcs-connect/vcs-connect /usr/bin/vcs-connect

WORKDIR /home/effxhq
USER 3339:3339

ENTRYPOINT [ "/usr/bin/vcs-connect" ]
