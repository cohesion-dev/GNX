#!/bin/bash
set -e

reflex -d none -s \
    -R 'auditlog/' \
    -R 'docs/' \
    -R 'tmp/' \
    -R '.*/volumes/.*' \
    -R '\.data' \
    -R '\.github' \
    -R '^coverage' \
    -R 'Makefile' \
    -R '.log$' \
    -R '_test.go$'\
    -- go run -trimpath cmd/main.go