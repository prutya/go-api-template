#!/bin/bash

set -exo pipefail

AIR_VERSION=v1.63.1

go install "github.com/air-verse/air@$AIR_VERSION"
