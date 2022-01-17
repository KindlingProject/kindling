#!/bin/bash

kubectl delete ds kindling-agent -n kindling
kubectl delete cm kindlingcfg -n kindling
kubectl delete -f kindling-clusterrole.yml
kubectl delete serviceaccounts kindling-agent -n kindling
kubectl delete clusterrolebindings kindling-agent