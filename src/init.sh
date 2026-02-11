#!/usr/bin/env bash
set -euo pipefail

cd /opt/OCPAlertGateway/src

if [ ! -f go.mod ]; then
  go mod init ocp-alert-gateway
fi

go get gopkg.in/yaml.v3
go mod tidy

