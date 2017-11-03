#!/bin/bash  
clear
echo "building coordinator election daemon..."
go build  
echo "<<< first use case >>"
echo "clear log"
rm --f msglog
echo ""
echo "starting daemons..."
#coordinator-election -admPort=8001 &
#coordinator-election -admPort=8002 &
#coordinator-election -admPort=8003 &
#coordinator-election -admPort=8004 &
#coordinator-election -admPort=8005 &
for i in {8001..8005}
do
   coordinator-election -admPort=$i &
done
sleep 0.5
echo ""
echo "starting coordinator..."
curl "http://localhost:8001?cmd=caste&PId=1&Coordinator=1&CId=2&HCId=2&SingleIP=2"
echo ""
echo "starting workers..."
curl "http://localhost:8002?cmd=caste&PId=2&Coordinator=1&CId=2&HCId=2&SingleIP=2"
curl "http://localhost:8003?cmd=caste&PId=3&Coordinator=1&CId=1&HCId=2&SingleIP=2"
curl "http://localhost:8004?cmd=caste&PId=4&Coordinator=1&CId=1&HCId=2&SingleIP=2"
curl "http://localhost:8005?cmd=caste&PId=5&Coordinator=1&CId=1&HCId=2&SingleIP=2"
echo ""
echo "running simulation"
curl "http://localhost:8002?cmd=casteCheckCoordinator"
curl -s "http://localhost:8001?cmd=stop"
curl "http://localhost:8002?cmd=casteCheckCoordinator"
curl "http://localhost:8003?cmd=casteCheckCoordinator"
curl -s "http://localhost:8002?cmd=stop"
curl "http://localhost:8003?cmd=casteCheckCoordinator"
curl "http://localhost:8004?cmd=casteCheckCoordinator"
sleep 0.1
curl "http://localhost:8004?cmd=casteCheckCoordinator"
echo ""
echo "stopping daemons"
curl -s "http://localhost:8003?cmd=stop"
curl -s "http://localhost:8004?cmd=stop"
curl -s "http://localhost:8005?cmd=stop"