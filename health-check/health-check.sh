#!/bin/bash

MESAGE_TO_SEND="HEALTCHECK"

. config-file


while true
do
  response=$(echo $MESAGE_TO_SEND | nc -w 5 $SERVER_IP $SERVER_PORT)
  timestamp=$(date +"%F %T")
  if [[ $response == $MESAGE_TO_SEND ]]
  then
    echo "$timestamp INFO     action: health-check | result: OK"
  else
    echo "$timestamp INFO     action: health-check | result: ERROR"
  fi

  sleep $PERIOD
done
