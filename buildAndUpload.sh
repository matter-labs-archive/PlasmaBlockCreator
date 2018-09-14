#!/bin/sh

docker build -t go-plasma -f Dockerfile_standalone --no-cache .
docker tag go-plasma thematterio/plasma:standalone-test
docker push thematterio/plasma:standalone-test