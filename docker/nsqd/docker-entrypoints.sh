#!/bin/sh

if [ "$1" = "sh" ];then
    sh
else
    exec /nsqd --lookupd-tcp-address=127.0.0.1:4160 --broadcast-address="${POD_IP}"
fi