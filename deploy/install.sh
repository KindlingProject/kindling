#!/bin/bash
kubectl create ns kindling

kubectl create serviceaccount kindling-agent -nkindling
kubectl apply -f kindling-clusterrole.yml
kubectl create clusterrolebinding kindling-agent --clusterrole=kindling-agent --serviceaccount=kindling:kindling-agent
kubectl create cm kindlingcfg -n kindling --from-file=kindling-collector-config.yml
kubectl apply -f kindling-deploy.yml

# configure Prometheus
kubectl apply -f kindling-prometheus-rabc.yml
kubectl apply -f kindling-prometheus-service.yml
kubectl apply -f kindling-prometheus-servicemonitor.yml

kubectl get pods -n kindling