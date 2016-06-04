#!/bin/bash -x

tc qdisc add dev eth0 handle 1: root htb

function delay {
  tc class add dev eth0 parent 1: classid 1:$3 htb rate 1000Mbps
  tc qdisc add dev eth0 parent 1:$3 handle $3: netem delay $2
  tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip sport $1 0xffff flowid 1:$3
}

function drop_and_delay {
  tc class add dev eth0 parent 1: classid 1:$4 htb rate 1000Mbps
  tc qdisc add dev eth0 parent 1:$4 handle $4: netem delay $2
  tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip sport $1 0xffff flowid 1:$4
  iptables -A INPUT  --dport $1 -m statistic --mode random --probability $3 -j DROP
  iptables -A OUTPUT --sport $1 -m statistic --mode random --probability $3 -j DROP
}

delay 8001 100ms 11
delay 8002 500ms 12
delay 8003 1s 13

delay 8006 100ms 16
delay 8007 500ms 17
delay 8008 1s 18

drop_and_delay 8010 500ms 0.2 20
drop_and_delay 8015 500ms 0.2 25

sudo -u nobody /main
