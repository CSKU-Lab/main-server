#!/bin/sh

source ./scripts/env_loader.sh

load_env_context .

atlas schema apply \
	--url $DATABASE_URL \
	--to "file://atlas/schema.hcl"
