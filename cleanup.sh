#!/bin/bash

# Delete all resources in the default namespace
kubectl delete deployments --all
kubectl delete services --all
kubectl delete pods --all
kubectl delete configmaps --all
kubectl delete secrets --all
kubectl delete ingresses --all
kubectl delete daemonsets --all
kubectl delete statefulsets --all
kubectl delete replicasets --all
kubectl delete jobs --all
kubectl delete cronjobs --all

# Optionally, delete resources across all namespaces
kubectl delete deployments --all --all-namespaces
kubectl delete services --all --all-namespaces
kubectl delete pods --all --all-namespaces
kubectl delete configmaps --all --all-namespaces
kubectl delete secrets --all --all-namespaces
kubectl delete ingresses --all --all-namespaces
kubectl delete daemonsets --all --all-namespaces
kubectl delete statefulsets --all --all-namespaces
kubectl delete replicasets --all --all-namespaces
kubectl delete jobs --all --all-namespaces
kubectl delete cronjobs --all --all-namespaces

echo "Cleanup completed."

