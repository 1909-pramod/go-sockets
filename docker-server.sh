#!/bin/bash
docker build . -t go-socket
docker-compose -f docker-compose.yml up