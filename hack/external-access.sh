#!/bin/bash

DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKDIR="$DIR/../k8s/external-access"
kubectl apply -f "$WORKDIR/cert-manager.yaml"
kubectl apply -f "$WORKDIR/awa-api-ingress.yaml"
kubectl apply -f "$WORKDIR/awa-api-redirect.yaml"
kubectl apply -f "$WORKDIR/awa-api-web.yaml"
kubectl apply -f "$WORKDIR/awa-api-certificate.yaml"
kubectl apply -f "$WORKDIR/awa-api-ingressroute.yaml"