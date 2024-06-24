#!/usr/bin/env bash
set -a && source .env && set +a && export GOOSE_DRIVER=postgres && export GOOSE_DBSTRING=$(echo $DB_URL | cut -d "?" -f1)