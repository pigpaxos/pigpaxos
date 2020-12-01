#!/usr/bin/env bash

source common.sh

DATE_WITH_TIME=`date "+%Y%m%d-%H%M%S"`
mkdir "/home/pigpaxos/paxi_logs/serv$DATE_WITH_TIME"

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
  download ${nodes[$i-1]} "/home/ubuntu/paxi/bin/logs/*" "/home/pigpaxos/paxi_logs/serv$DATE_WITH_TIME/"
  echo "Downloaded logs from 1.$i"
done