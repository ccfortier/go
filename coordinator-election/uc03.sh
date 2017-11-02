#!/bin/bash  
clear
echo "building coordinator election daemon..."
go build  
echo "<<< third use case >>"
echo "clear log"
rm --f msglog
echo ""
echo "starting daemons..."
coordinator-election -admPort=8001 &
coordinator-election -admPort=8002 &
coordinator-election -admPort=8003 &
coordinator-election -admPort=8004 &
coordinator-election -admPort=8005 &
coordinator-election -admPort=8006 &
coordinator-election -admPort=8007 &
coordinator-election -admPort=8008 &
coordinator-election -admPort=8009 &
sleep 0.1
echo ""
echo "starting coordinator..."
curl "http://localhost:8001?cmd=caste&PId=1&Coordinator=1&CId=3&HCId=3&SingleIP=2"
echo ""
echo "starting workers..."
curl "http://localhost:8002?cmd=caste&PId=2&Coordinator=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:8003?cmd=caste&PId=3&Coordinator=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:8004?cmd=caste&PId=4&Coordinator=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:8005?cmd=caste&PId=5&Coordinator=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:8006?cmd=caste&PId=6&Coordinator=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:8007?cmd=caste&PId=7&Coordinator=1&CId=1&HCId=3&SingleIP=2"
curl "http://localhost:8008?cmd=caste&PId=8&Coordinator=1&CId=1&HCId=3&SingleIP=2"
curl "http://localhost:8009?cmd=caste&PId=9&Coordinator=1&CId=1&HCId=3&SingleIP=2"
echo ""
echo "running simulation"
curl "http://localhost:8002?cmd=casteCheckCoordinator"
curl -s "http://localhost:8001?cmd=stop"
curl "http://localhost:8008?cmd=casteCheckCoordinator"
sleep 0.5
curl "http://localhost:8002?cmd=casteDump"
curl "http://localhost:8003?cmd=casteDump"
curl "http://localhost:8004?cmd=casteDump"
curl "http://localhost:8005?cmd=casteDump"
curl "http://localhost:8006?cmd=casteDump"
curl "http://localhost:8007?cmd=casteDump"
curl "http://localhost:8008?cmd=casteDump"
curl "http://localhost:8009?cmd=casteDump"
echo ""
echo "stopping daemons"
curl -s "http://localhost:8002?cmd=stop"
curl -s "http://localhost:8003?cmd=stop"
curl -s "http://localhost:8004?cmd=stop"
curl -s "http://localhost:8005?cmd=stop"
curl -s "http://localhost:8006?cmd=stop"
curl -s "http://localhost:8007?cmd=stop"
curl -s "http://localhost:8008?cmd=stop"
curl -s "http://localhost:8009?cmd=stop"