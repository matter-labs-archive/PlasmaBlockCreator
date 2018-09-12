#!/bin/bash
cd redisPrep && npm install && node prepareRedis.js && cd ..
go run -v server.go