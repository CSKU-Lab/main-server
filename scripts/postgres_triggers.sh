#!/bin/sh

SQL_FILE="./postgresql/submission_outbox_trigger.sql"

# 1. Check if the file exists
if [ ! -f "$SQL_FILE" ]; then
    echo "Error: File '$SQL_FILE' not found."
    exit 1
fi

POSTGRES_CONTAINER_ID=$(docker ps | grep "postgres" | awk '{print $1}')

cat $SQL_FILE | docker exec -i $POSTGRES_CONTAINER_ID psql -U $POSTGRES_USER -d $MAIN_SERVER_DATABASE_NAME --set ON_ERROR_STOP=1

if [ $? -eq 0 ]; then
    echo "Successfully executed $SQL_FILE"
else
    echo "An error occurred during SQL execution."
    exit 1
fi
