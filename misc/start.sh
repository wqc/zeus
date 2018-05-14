#!/bin/bash

dir=`dirname $0`
cd $dir/..
path=`pwd`
process=$1

mkdir -p logs
out="logs/$process.out"

trap '' 1 2 3

for (( ;; ))
do
    $path/bin/$process >> $out 2>&1

	echo "`date +%Y%m%d%H%M%S` $process exit" >> $out
    sleep 5
done

