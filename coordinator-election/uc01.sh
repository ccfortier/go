#!/bin/bash  
clear
echo "building coordinator election daemon..."
go build  
echo "first use case"
coordinator-election -admPort=8002 &
coordinator-election -admPort=8003 &
curl "http://localhost:8002?cmd=caste&PId=2&Coordinator=2&SingleIP=2"
curl "http://localhost:8003?cmd=caste&PId=3&Coordinator=2&SingleIP=2"
curl -s "http://localhost:8002?cmd=stop"
curl "http://localhost:8003?cmd=casteCheckCoordinator"
curl -s "http://localhost:8003?cmd=stop"