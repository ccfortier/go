#!/bin/bash  
clear
echo "clear log"
rm --f msglog
echo "building coordinator election daemon..."
go build  
echo "<<< third use case >>"
echo ""
echo "starting daemons..."
for i in {8001..8015}
do
   coordinator-election -admPort=$i &
done
sleep 0.1
echo ""
echo "starting coordinator on caste 3..."
curl "http://localhost:8001?cmd=caste&PId=1&Coordinator=1&CId=3&HCId=3&SingleIP=2"
echo ""
echo "starting workers on caste 2..."
for i in {2..6}
do
   curl "http://localhost:$((8000 + $i))?cmd=caste&PId=$i&Coordinator=1&CId=2&HCId=3&SingleIP=2"
done
echo ""
echo "starting workers on caste 1..."
for i in {7..15}
do
   curl "http://localhost:$((8000 + $i))?cmd=caste&PId=$i&Coordinator=1&CId=1&HCId=3&SingleIP=2"
done
echo ""
echo "running simulation"
curl "http://localhost:8002?cmd=casteCheckCoordinator"
curl -s "http://localhost:8001?cmd=stop"
curl "http://localhost:8007?cmd=casteCheckCoordinator"
sleep 0.1
for i in {8002..8015}
do
   curl "http://localhost:$i?cmd=casteDump"
done
echo ""
echo "stopping daemons"
for i in {8002..8015}
do
   curl -s "http://localhost:$i?cmd=sStop"
done
echo ""
echo "showing message log"
cat msglog