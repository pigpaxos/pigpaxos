#!/usr/bin/env bash

PID_FILE=server.pid

PID=$(cat "${PID_FILE}");

if [ -z "${PID}" ]; then
    echo "Process id for servers is written to location: {$PID_FILE}"
    go build ../server/
    go build ../client/
    go build ../cmd/
    rm -r logs
    mkdir logs/
    ./server -log_dir=logs -log_level=info -id 1.1 -algorithm=epaxos>logs/out1.1.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.2 -algorithm=epaxos>logs/out1.2.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.3 -algorithm=epaxos>logs/out1.3.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.4 -algorithm=epaxos>logs/out1.4.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.5 -algorithm=epaxos>logs/out1.5.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.6 -algorithm=epaxos>logs/out1.6.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.7 -algorithm=epaxos>logs/out1.7.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.8 -algorithm=epaxos>logs/out1.8.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=info -id 1.9 -algorithm=epaxos>logs/out1.9.txt 2>&1 &
    echo $! >> ${PID_FILE}
else
    echo "Servers are already started in this folder."
    exit 0
fi
