#!/bin/bash  
clear
echo "clear log"
rm --f msglog
echo "building leader election daemon..."
go build  
echo "<<< third use case >>"
echo ""
echo "starting daemons..."
for i in {10001..11500}
do
   ./leader-election -admPort=$i -quiet &
done
sleep 0.1
echo ""
echo "starting leader on caste 3..."
curl "http://localhost:10001?cmd=caste&PId=1&Leader=1&CId=3&HCId=3&SingleIP=2"
echo ""
echo "starting 99 workers on caste 2..."
for i in {2..100}
do
   curl "http://localhost:$((10000 + $i))?cmd=caste&PId=$i&Leader=1&CId=2&HCId=3&SingleIP=2"
done
echo ""
echo "starting 1400 workers on caste 1..."
for i in {101..1500}
do
   curl "http://localhost:$((10000 + $i))?cmd=caste&PId=$i&Leader=1&CId=1&HCId=3&SingleIP=2"
done
echo ""
echo "running simulation"
curl "http://localhost:10002?cmd=casteCheckLeader"
curl -s "http://localhost:10001?cmd=stop"
curl "http://localhost:10101?cmd=casteCheckLeader"
sleep 0.1
for i in {10002..11500}
do
   curl "http://localhost:$i?cmd=casteDump"
done
echo ""
echo "stopping daemons"
for i in {10002..11500}
do
   curl -s "http://localhost:$i?cmd=sStop"
done
echo ""
echo "showing message log"
cat msglog