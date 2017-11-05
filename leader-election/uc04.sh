#!/bin/bash  
clear
echo "building coordinator election daemon..."
go build  
echo "<<< fourth use case >>"
echo "clear log"
rm --f msglog
echo ""
echo "starting daemons..."
for i in {10001..10009}
do
   ./leader-election -admPort=$i &
done
sleep 0.1
echo ""
echo "starting coordinator at caste 3..."
curl "http://localhost:10001?cmd=caste&PId=1&Leader=1&CId=3&HCId=3&SingleIP=2"
echo ""
echo "starting workers at caste 3..."
curl "http://localhost:10002?cmd=caste&PId=2&Leader=1&CId=3&HCId=3&SingleIP=2"
curl "http://localhost:10003?cmd=caste&PId=3&Leader=1&CId=3&HCId=3&SingleIP=2"
echo ""
echo "starting workers at caste 2..."
curl "http://localhost:10004?cmd=caste&PId=4&Leader=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:10005?cmd=caste&PId=5&Leader=1&CId=2&HCId=3&SingleIP=2"
curl "http://localhost:10006?cmd=caste&PId=6&Leader=1&CId=2&HCId=3&SingleIP=2"
echo ""
echo "starting workers at caste 1..."
curl "http://localhost:10007?cmd=caste&PId=7&Leader=1&CId=1&HCId=3&SingleIP=2"
curl "http://localhost:10008?cmd=caste&PId=8&Leader=1&CId=1&HCId=3&SingleIP=2"
curl "http://localhost:10009?cmd=caste&PId=9&Leader=1&CId=1&HCId=3&SingleIP=2"
echo ""
echo "running simulation"
curl "http://localhost:10002?cmd=casteCheckLeader"
curl -s "http://localhost:10001?cmd=stop"
curl -s "http://localhost:10004?cmd=stop"
curl -s "http://localhost:10005?cmd=stop"
curl -s "http://localhost:10006?cmd=stop"
curl "http://localhost:10008?cmd=casteCheckLeader"
sleep 0.1
for i in {10002..10003}
do
   curl "http://localhost:$i?cmd=casteDump"
done
for i in {10007..10009}
do
   curl "http://localhost:$i?cmd=casteDump"
done
echo ""
echo "stopping daemons"
for i in {10002..10003}
do
   curl -s "http://localhost:$i?cmd=sStop"
done
for i in {10007..10009}
do
   curl -s "http://localhost:$i?cmd=sStop"
done
echo ""
echo "showing message log"
cat msglog