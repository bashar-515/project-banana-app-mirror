#!/bin/sh

echo "PB_SERVER_PORT=${PB_SERVER_PORT}" > .env
echo "PB_SERVER_HOST=${PB_SERVER_HOST}" >> .env

exec ./server
