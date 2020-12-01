#!/usr/bin/env bash

source common.sh

client
for (( i=30; i>0; i-- ));
do
  echo "$i"
  sleep 1
done
