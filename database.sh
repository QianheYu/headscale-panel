#!/bin/bash

# Set the PostgreSQL connection parameters
HOST="localhost"
PORT="5432"
USERNAME="postgres"
DATABASE="postgres"

while getopts ":h:P:u:f:p:d:" opt; do
  case $opt in
    h) HOST="$OPTARG"
 ;;
    P) PORT="$OPTARG"
    ;;
    u) USERNAME="$OPTARG"
    ;;
    p) PASSWORD="$OPTARG"
    ;;
    d) DATABASE="$OPTARG"
    ;;
    f) FILE="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done


# Connect to the PostgreSQL database and execute the SQL query
psql -h $HOST -p $PORT -U $USERNAME -d $DATABASE -f $FILE
