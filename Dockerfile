ARG ALPINE_VERSION
ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as base

# Install golangci-lint
ARG LINTER_VERSION
RUN apk update --no-cache && \
	apk add make bash g++ git golangci-lint=~${LINTER_VERSION}

# Set up workspace
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Install go-test-coverage
ARG COVER_VERSION
RUN go install github.com/vladopajic/go-test-coverage/v2@${COVER_VERSION}

# Install mockery
ARG MOCKERY_VERSION
RUN go install github.com/vektra/mockery/v2@${MOCKERY_VERSION}

# Install ginkgo
RUN go install github.com/onsi/ginkgo/v2/ginkgo


FROM base as build

RUN mkdir bin
COPY . .

# Build the server
ARG BUILD_VERSION=0.0.0
ARG BUILD_DATE=YYYYMMDD
ARG BUILD_COMMIT=dev
RUN go build -o ./bin/server \
	-ldflags " \
		-X 'main.Version=${BUILD_VERSION}' \
		-X 'main.Date=${BUILD_DATE}' \
		-X 'main.Commit=${BUILD_COMMIT}' \
	" ./server


FROM alpine:${ALPINE_VERSION}

# Install the server binary
COPY --from=build /workspace/bin/server /usr/local/bin/server

# Start the server
CMD server
