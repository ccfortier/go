#!/bin/bash  
echo "first use case"  
coordinator-election -admPort=8002 &
curl "http://localhost:8002?cmd=caste&PId=2&Coordinator=2&SingleIP=2"
coordinator-election -admPort=8003 &
curl "http://localhost:8002?cmd=caste&PId=3&Coordinator=2&SingleIP=2"
curl "http://localhost:8002?cmd=stop"
curl "http://localhost:8003?cmd=casteCoordinator"
curl "http://localhost:8003?cmd=stop"