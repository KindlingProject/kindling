#!/bin/bash
kubectl create ns kindling

kubectl create serviceaccount kindling-agent -nkindling
kubectl apply -f kindling-clusterrole.yml
kubectl apply -f kindling-clusterrolebinding.yml
kubectl create cm kindlingcfg -n kindling --from-file=kindling-collector-config.yml
kubectl apply -f kindling-deploy.yml

# configure Prometheus
kubectl apply -f kindling-prometheus-rabc.yml
kubectl apply -f kindling-prometheus-service.yml
kubectl apply -f kindling-prometheus-servicemonitor.yml
sleep 5
kubectl get pods -n kindling