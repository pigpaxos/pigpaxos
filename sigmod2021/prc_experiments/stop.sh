#!/usr/bin/env bash

source common.sh

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
  ssh -i va.pem ubuntu@${nodes[$i-1]} 'cd /home/ubuntu/paxi/bin; ./stop.sh'
  echo "Stopped 1.$i"
done

