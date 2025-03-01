#!/bin/bash

# Run integration tests
cd "$(dirname "$0")"
go test -v .
