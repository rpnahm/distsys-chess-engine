#!/bin/bash

# run a series of tests on the test servers

#$ -M rnahm@nd.edu   # Email address for job notification
#$ -m abe            # Send mail when job begins, ends and aborts
#$ -pe smp 2         # Specify parallel environment and legal core size
#$ -q long           # Specify queue
#$ -N short-c      # Specify job name


module use -a ~/privatemodules
module load golang/1.20      # Required modules

 # Application to execute
cd ~/distsys-chess-engine

# running with between 1 and 16 cores for 10 games with 1000ms turnTime, and a single threaded engine
./bin/test test 1  1000 10 1
./bin/test test 2  1000 10 1
./bin/test test 4  1000 10 1
./bin/test test 8  1000 10 1
./bin/test test 16 1000 10 1
