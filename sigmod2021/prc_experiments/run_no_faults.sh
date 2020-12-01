#!/usr/bin/env bash

./upload_config.sh config_bigpaxos25_10.json config.json

echo "No faults, 3 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 3 0 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_0_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 3 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 3 1 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_1_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 3 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 3 2 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_2_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 3 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 3 3 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 3_3_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

# 2 Relay Groups

echo "No faults, 2 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 2 0 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_0_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 2 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 2 1 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_1_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 2 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 2 2 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_2_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 2 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 2 3 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 2_3_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

# 1 Relay Group

echo "No faults, 1 relay groups, 0 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 1 0 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_0_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 1 relay groups, 1 node slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 1 1 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_1_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 1 relay groups, 2 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 1 2 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_2_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 1 relay groups, 3 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 1 3 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_3_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done

echo "No faults, 1 relay groups, 12 nodes slack"
for i in {1..5}
do
   echo "Experiment Run # $i starting"
   ./start_no_faults.sh 1 12 # starts servers
   ./bench.sh # run 60 second benchmark
   sleep 10 # give a bit more slack, just in case
   ./stop.sh  # stop servers
   ./get_logs.sh 1_12_nf_
    echo "Experiment Run # $i Done"
    sleep 1
done