############################
# Builder
############################
FROM ubuntu:24.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=en_US.utf8
ENV TZ=Asia/Seoul

# Base packages (builder)
RUN apt-get update -q && \
    apt-get install --no-install-recommends -y -q \
      ca-certificates tzdata locales \
      curl wget git \
      build-essential && \
    apt-get clean -y -q && \
    rm -rf /var/lib/apt/lists/* && \
    localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install Go (latest)
RUN GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1) && \
    wget -nv https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf ${GO_VERSION}.linux-amd64.tar.gz && \
    rm ${GO_VERSION}.linux-amd64.tar.gz

ENV GOROOT=/usr/local/go \
    GOPATH=/app \
    PATH=$PATH:/usr/local/go/bin:/app/bin

WORKDIR /app

# Build args (override when needed)
ARG REPO_URL=https://github.com/rootsnet/OCPAlertGateway.git
ARG REF=main

# Clone repository
RUN git clone --depth 1 --branch ${REF} ${REPO_URL} /app/OCPAlertGateway

# Build from src/ (your project keeps Go sources under src/)
WORKDIR /app/OCPAlertGateway/src

# If go.mod does not exist, initialize module and fetch minimal dependency.
# Then run go mod tidy and build the binary.
RUN if [ ! -f go.mod ]; then \
      go mod init ocp-alert-gateway && \
      go get gopkg.in/yaml.v3 ; \
    fi && \
    go mod tidy && \
    mkdir -p /app/bin && \
    go build -trimpath -ldflags "-s -w" -o /app/bin/ocp-alert-gateway .

############################
# Runtime
############################
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=en_US.utf8
ENV TZ=Asia/Seoul

# Minimal runtime dependencies
RUN apt-get update -q && \
    apt-get install --no-install-recommends -y -q \
      ca-certificates tzdata locales && \
    apt-get clean -y -q && \
    rm -rf /var/lib/apt/lists/* && \
    localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# App layout
WORKDIR /app
RUN mkdir -p /app/bin /app/config

# Copy binary from builder
COPY --from=builder /app/bin/ocp-alert-gateway /app/bin/ocp-alert-gateway

ENV PATH="/app/bin:$PATH"

EXPOSE 8080

# Expect /app/config/config.yaml to be mounted/provided at runtime
CMD ["ocp-alert-gateway", "-config", "/app/config/config.yaml"]
