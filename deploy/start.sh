#!/bin/bash
kubectl create ns kindling

kubectl create serviceaccount kindling-agent -nkindling
kubectl apply -f https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-clusterrole.yml
kubectl create clusterrolebinding kindling-agent --clusterrole=kindling-agent --serviceaccount=kindling:kindling-agent
curl -O https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-collector-config.yml
kubectl create cm kindlingcfg -n kindling --from-file=kindling-collector-config.yml
kubectl apply -f https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-deploy.yml

# configure Prometheus
kubectl apply -f https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-prometheus-rabc.yml
kubectl apply -f https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-prometheus-service.yml
kubectl apply -f https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/kindling-prometheus-servicemonitor.yml

kubectl get pods -n kindling