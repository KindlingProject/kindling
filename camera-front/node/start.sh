#!/bin/sh
 cd `dirname $0`
 pids=$(netstat -ltnp|grep 9900|awk '{print $7}');
 if [ $pids ] ; then
  pm2-runtime restart camera-node         
 else
  pm2-runtime start index.js --name camera-node 
 fi 
#  pids=$(netstat -ltnp|grep 9091|awk '{print $7}');
#  if [ $pids ] ; then
#   pm2 restart apm-grafana-proxy       
#  else
#   pm2 start grafana.js --name apm-grafana-proxy
#  fi 
