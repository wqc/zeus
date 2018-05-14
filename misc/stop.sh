#!/bin/bash

dir=`dirname $0`
cd $dir/..
path=`pwd`
process=$1

n=0
for (( ;; ))
do

	pid=`ps -ef | grep $process | grep -v stop.sh | grep -v grep | awk -F ' ' '{print $2}'`

	if [ "$pid" = "" ]; then
		echo "$process exited"
		exit 0
	fi

	if [ $n -eq 5 ]; then
		echo "stop $process faild"
		exit -1
	fi

	echo "$process stopping ..."
	kill $pid
    sleep 1

	n=$((n+1))
done
