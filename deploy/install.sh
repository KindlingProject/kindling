#!/bin/bash

kubectl create serviceaccount kindling-agent -nkindling
kubectl apply -f kindling-clusterrole.yml
kubectl create clusterrolebinding kindling-agent --clusterrole=kindling-agent --serviceaccount=kindling:kindling-agent
kubectl apply -f kindling-configmap.yml
kubectl apply -f kindling-deploy.yml

kubectl get pods -n kindling