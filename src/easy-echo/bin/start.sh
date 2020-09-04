#!/bin/bash
cd `dirname $0`
cd ../
killall easy-echo >/dev/null 2>&1
nohup ./bin/easy-echo >> log/console_output.log 2>&1 &
