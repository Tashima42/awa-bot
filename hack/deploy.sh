#!/bin/bash

DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKDIR="$DIR/../k8s/bot"
cd "$DIR/.." 
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
git pull
kubectl apply -f "$WORKDIR/database-persistent-volume-claim.yaml"
kubectl apply -f "$WORKDIR/postgres-deployment.yaml"
kubectl apply -f "$WORKDIR/postgres-cluster-ip-service.yaml"
kubectl apply -f "$WORKDIR/bot-deployment.yaml"
kubectl apply -f "$WORKDIR/api-deployment.yaml"
kubectl apply -f "$WORKDIR/api-service.yaml"
kubectl apply -f "$WORKDIR/admin-deployment.yaml"
kubectl apply -f "$WORKDIR/admin-service.yaml"
kubectl patch deployment bot-deployment -p "{\"spec\": {\"template\": {\"metadata\": { \"labels\": {  \"redeploy\": \"$(date +%s)\"}}}}}"
kubectl patch deployment api-deployment -p "{\"spec\": {\"template\": {\"metadata\": { \"labels\": {  \"redeploy\": \"$(date +%s)\"}}}}}"
kubectl patch deployment admin-deployment -p "{\"spec\": {\"template\": {\"metadata\": { \"labels\": {  \"redeploy\": \"$(date +%s)\"}}}}}"
$DIR/external-access.sh
