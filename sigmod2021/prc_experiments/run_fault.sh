#!/usr/bin/env bash

./upload_config.sh config_bigpaxos25_10.json config.json

echo "One fault, 3 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 3 0 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_0_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 3 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 3 1 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_1_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 3 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 3 2 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_2_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 3 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 3 3 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_3_
    echo "Experiment Run # $i Done"
    sleep 1
done

# 2 Relay Groups

echo "One fault, 2 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 2 0 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_0_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 2 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 2 1 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_1_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 2 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 2 2 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_2_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 2 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 2 3 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_3_
    echo "Experiment Run # $i Done"
    sleep 1
done

# 1 Relay Group

echo "One fault, 1 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 1 0 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_0_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 1 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 1 1 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_1_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 1 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 1 2 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_2_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 1 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 1 3 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_3_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "One fault, 1 relay groups, 12 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_1_fault.sh 1 12 # starts servers
   ./bench.sh # run 30 second benchmark
   sleep5 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_12_
    echo "Experiment Run # $i Done"
    sleep 1
done