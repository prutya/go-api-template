#!/bin/bash

set -exo pipefail

AIR_VERSION=v1.62.0

go install "github.com/air-verse/air@$AIR_VERSION"
