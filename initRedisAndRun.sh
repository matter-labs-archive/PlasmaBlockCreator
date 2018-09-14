#!/bin/bash
fdbcli --exec "status details"
cd redisPrep && npm install && node prepareRedis.js && cd ..
go run -v server.go