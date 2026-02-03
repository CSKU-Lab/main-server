#!/bin/sh

stamp_file="trigger_stamp.json"

# Initialize the JSON file as an empty array if it doesn't exist
if [ ! -f "$stamp_file" ]; then
    echo "[]" > "$stamp_file"
fi

POSTGRES_CONTAINER_ID=$(docker ps | grep "postgres" | awk '{print $1}')

# Use 'find' to list files and 'read' to loop through them safely
find ./postgresql/ -maxdepth 2 -type f | while read -r file; do
    if [ "$file" != "$stamp_file" ]; then
        already_exists=$(jq --arg f "$file" 'contains([$f])' "$stamp_file")

        if [ "$already_exists" == "true" ]; then
            echo "Skipping $file: already triggered."
            continue
        fi
    fi
    echo "Apply $file to the postgresql"
    cat $file | docker exec -i $POSTGRES_CONTAINER_ID "psql -U $POSTGRES_USER -d $MAIN_SERVER_DATABASE_NAME --set ON_ERROR_STOP=1" > /dev/null
    tmp=$(jq --arg f "$file" '. += [$f]' "$stamp_file")
    echo "$tmp" > "$stamp_file"
done
