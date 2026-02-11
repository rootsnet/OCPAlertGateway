#!/usr/bin/env bash
set -euo pipefail

cd /opt/OCPAlertGateway/src

go build -o /opt/OCPAlertGateway/bin/ocp-alert-gateway .
