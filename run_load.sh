#!/bin/bash
GOMAXPROCS=4 GOCACHE=off go test -v loadTest/createAndSpend_test.go