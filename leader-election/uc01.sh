#!/bin/bash  
clear
echo "building leader-election daemon..."
go build  
echo "<<< first use case >>"
echo "clear log"
rm --f msglog
echo ""
echo "starting daemons..."
for i in {10001..10008}
do
   ./leader-election -admPort=$i &
done
sleep 0.5
echo ""
echo "starting leader on caste 2..."
curl "http://localhost:10001?cmd=caste&PId=1&Leader=1&CId=2&HCId=2&SingleIP=2"
echo
echo "starting workers on caste 1..."
for i in {2..8}
do
   curl "http://localhost:$((10000 + $i))?cmd=caste&PId=$i&Leader=1&CId=1&HCId=2&SingleIP=2"
done
echo ""
echo "running simulation"
curl "http://localhost:10002?cmd=casteCheckLeader"
curl -s "http://localhost:10001?cmd=stop"
curl "http://localhost:10002?cmd=casteCheckLeader"
curl "http://localhost:10003?cmd=casteCheckLeader"
sleep 0.1
sleep 0.1
for i in {10003..10008}
do
   curl "http://localhost:$i?cmd=casteDump"
done
curl -s "http://localhost:10002?cmd=stop"
curl "http://localhost:10003?cmd=casteCheckLeader"
sleep 0.1
for i in {10003..10008}
do
   curl "http://localhost:$i?cmd=casteDump"
done
echo ""
echo "stopping daemons"
for i in {10003..10008}
do
   curl -s "http://localhost:$i?cmd=sStop"
done
echo ""
echo "showing message log"
cat msglog