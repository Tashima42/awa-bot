#!/bin/bash

eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
git pull
sudo k3s kubectl apply -f ./k8s
kubectl patch deployment bot-deployment -p "{\"spec\": {\"template\": {\"metadata\": { \"labels\": {  \"redeploy\": \"$(date +%s)\"}}}}}"