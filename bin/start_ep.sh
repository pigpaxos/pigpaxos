#!/usr/bin/env bash

PID_FILE=server.pid

./server -log_dir=logs -log_level=info -id $1 -algorithm $2 > logs/out1.$1.txt 2>&1 &
echo $! >> ${PID_FILE}
