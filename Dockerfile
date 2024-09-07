ARG ALPINE_VERSION
ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS base

# Install golangci-lint, protoc
ARG LINTER_VERSION
ARG PROTOC_VERSION
RUN apk update --no-cache && \
	apk add make bash g++ git \
		golangci-lint=~${LINTER_VERSION} \
		protoc=~${PROTOC_VERSION}

# Set up dev user
ARG USER_ID=1000
ARG GROUP_ID=1000
RUN addgroup -g ${GROUP_ID} dev && \
	adduser -u ${USER_ID} -G dev -D dev && \
	mkdir -p /etc/sudoers.d && \
	echo "dev ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/dev && \
	chmod 0440 /etc/sudoers.d/dev && \
	mkdir /.cache && \
	chmod -R 777 /.cache
USER ${USER_ID}:${GROUP_ID}

# Install go binaries
ARG PROTOC_GO_VERSION
ARG PROTOC_GO_GRPC_VERSION
ARG COVER_VERSION
ARG MOCKERY_VERSION
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GO_VERSION} && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GO_GRPC_VERSION} && \
	go install github.com/vladopajic/go-test-coverage/v2@${COVER_VERSION} && \
	go install github.com/vektra/mockery/v2@${MOCKERY_VERSION}

# Set up workspace
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Install ginkgo (has to installed in the context of a workspace)
RUN go install github.com/onsi/ginkgo/v2/ginkgo

FROM base AS build

RUN mkdir bin
COPY . .

ARG BUILD_VERSION=0.0.0
ARG BUILD_DATE=YYYYMMDD
ARG BUILD_COMMIT=dev

# Build the server
RUN go build -o ./bin/server \
	-buildvcs=false \
	-ldflags " \
		-X 'main.Version=${BUILD_VERSION}' \
		-X 'main.Date=${BUILD_DATE}' \
		-X 'main.Commit=${BUILD_COMMIT}' \
	" ./server

# Build the client
RUN go build -o ./bin/cli ./cli


FROM alpine:${ALPINE_VERSION}

# Install the server binary
COPY --from=build /workspace/bin/server /usr/local/bin/server

# Install the cli binary
COPY --from=build /workspace/bin/cli /usr/local/bin/cli

# Start the server
CMD ["server"]
