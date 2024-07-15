#!/bin/bash

set -u
set -e
# set -x

# The Promtail daemonset runs in the default namespace.

NAMESPACE=default
SECRETS=carlo-comments-secrets
kubectl -n ${NAMESPACE} delete secret ${SECRETS} || echo "${SECRETS}"
kubectl create secret generic ${SECRETS} \
    --namespace=${NAMESPACE} \
    --from-literal=LOKI_URL=${LOKI_URL} \
    --from-literal=LOKI_USER=${LOKI_USER} \
    --from-literal=LOKI_API_TOKEN=${LOKI_API_TOKEN}
kubectl -n ${NAMESPACE} describe secret ${SECRETS}


NAMESPACE=backend
kubectl create namespace $NAMESPACE || echo $NAMESPACE
SECRETS=carlo-comments-secrets
kubectl -n ${NAMESPACE} delete secret ${SECRETS} || echo "${SECRETS}"
kubectl create secret generic ${SECRETS} \
    --namespace=${NAMESPACE} \
    --from-literal=AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
    --from-literal=AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    --from-literal=HMAC_SECRET=${HMAC_SECRET} \
    --from-literal=ADMIN_USER=${ADMIN_USER} \
    --from-literal=ADMIN_PASS=${ADMIN_PASS} \
    --from-literal=S3_REGION=${S3_REGION} \
    --from-literal=S3_BUCKET=${S3_BUCKET} \
    --from-literal=SES_IDENTITY=${SES_IDENTITY}
kubectl -n ${NAMESPACE} describe secret ${SECRETS}

REGISTRY=regcred
kubectl -n ${NAMESPACE} delete secret ${REGISTRY} || echo "${REGISTRY}"
kubectl create secret docker-registry ${REGISTRY} \
    --namespace=${NAMESPACE} \
    --docker-server=registry.digitalocean.com \
    --docker-username=carlo@carlo-hamalainen.net \
    --docker-password=${DO_K8S_TOKEN} \
    --docker-email=carlo@carlo-hamalainen.net
kubectl -n ${NAMESPACE} describe secret ${REGISTRY}


CONFIG_MAP=carlo-comments-config
kubectl -n ${NAMESPACE} delete configmap ${CONFIG_MAP} || echo "${CONFIG_MAP}"
kubectl create configmap ${CONFIG_MAP} \
    --namespace=${NAMESPACE} \
    --from-literal=PORT=${PORT} \
    --from-literal=APP_NAME=${APP_NAME} \
    --from-literal=LOG_DIRECTORY=${LOG_DIRECTORY} \
    --from-literal=HANDLER_TIMEOUT=${HANDLER_TIMEOUT} \
    --from-literal=LIMITER_RATE=${LIMITER_RATE} \
    --from-literal=LIMITER_BURST=${LIMITER_BURST} \
    --from-literal=COMMENT_HOST=${COMMENT_HOST} \
    --from-literal=CORS_ALLOWED_ORIGINS="https://carlo-hamalainen.net,http://localhost,http://localhost:8000,http://localhost:1313" # FIXME
kubectl -n ${NAMESPACE} describe configmap ${CONFIG_MAP}