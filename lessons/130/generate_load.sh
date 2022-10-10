#!/bin/bash

while true; do
  sql=$(cat <<EOF
  {"email": "$(openssl rand -hex 6)@gmail.com", "password": "$(openssl rand -hex 12)"})

  sleep 0.5 &
  curl -H "Content-Type:application/json" -d "$sql" http://localhost:8081/users
  curl -H "Content-Type:application/json" -d "$sql" http://localhost:8082/users
  
  echo ""
  wait
done
