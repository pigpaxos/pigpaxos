#!/usr/bin/env bash

source common.sh

DATE_WITH_TIME=`date "+%Y%m%d-%H%M%S"`
mkdir "/home/pigpaxos/paxi_logs/$1$DATE_WITH_TIME"
download $va_client "/home/ubuntu/paxi/bin/*.log" "/home/pigpaxos/paxi_logs/$1$DATE_WITH_TIME"
download $va_client "/home/ubuntu/paxi/bin/latency" "/home/pigpaxos/paxi_logs/$1$DATE_WITH_TIME"
download $va_client "/home/ubuntu/paxi/bin/history.csv" "/home/pigpaxos/paxi_logs/$1$DATE_WITH_TIME"
ssh -i va.pem ubuntu@$va_client 'cd /home/ubuntu/paxi/; rm bin/*.log'
ssh -i va.pem ubuntu@$va_client 'cd /home/ubuntu/paxi/bin/; rm logs/*.log'
echo "Downloaded logs from client"