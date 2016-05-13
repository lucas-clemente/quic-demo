#!/bin/bash -x

tc qdisc add dev eth0 handle 1: root htb
function delay {
  tc class add dev eth0 parent 1: classid 1:$3 htb rate 1000Mbps
  tc qdisc add dev eth0 parent 1:$3 handle $3: netem delay $2
  tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip sport $1 0xffff flowid 1:$3
}

delay 8001 100ms 11
delay 8002 200ms 12
delay 8003 500ms 13
delay 8004 1s 14

delay 8006 100ms 16
delay 8007 200ms 17
delay 8008 500ms 18
delay 8009 1s 19

sudo -u nobody /main
