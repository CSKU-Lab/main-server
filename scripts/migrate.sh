#!/bin/sh

DATABASE_URL=$(echo "$MAIN_SERVER_DATABASE_URL" | sed 's/db/localhost/g')

atlas schema apply \
	--url $DATABASE_URL \
	--to "file://atlas/schema.hcl" \
	"$@"
