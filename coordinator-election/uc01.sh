#!/bin/bash  
clear
echo "building coordinator election daemon..."
go build  
echo "first use case"
coordinator-election -admPort=8001 &
coordinator-election -admPort=8002 &
curl "http://localhost:8001?cmd=caste&PId=1&Coordinator=1&CId=2&HCId=2&SingleIP=2"
curl "http://localhost:8001?cmd=casteDump"
curl "http://localhost:8002?cmd=caste&PId=2&Coordinator=1&CId=1&HCId=2&SingleIP=2"
curl "http://localhost:8002?cmd=casteDump"
curl -s "http://localhost:8001?cmd=stop"
curl "http://localhost:8002?cmd=casteCheckCoordinator"
curl -s "http://localhost:8002?cmd=stop"