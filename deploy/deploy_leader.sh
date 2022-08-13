#!/bin/bash
keylocation='/Users/reedtrevelyan/Desktop/ec2_keys'
echo "Starting deployment."

die() {
  printf -- '%s\n' "$*"
  exit 1
}

if test -z "$PI_HOST"; then
  echo "Missing PI_HOST env variable, exiting."
  exit 1
fi

# first copy config file
scp -i "$keylocation/aucbot-orch-prod.pem" "$(pwd)/quillsecure_prod.yaml" "$PI_HOST":/home/quillsecure/quillsecure_prod.yaml || die "Failed to copy orch-config.prod.sh."

# then copy bot binary
scp "$(pwd)/cmd/orchestrator/orchestrator" "$PI_HOST":/usr/local/bin/leader_new || die "Failed to copy leader_new."

# then restart service, swapping in the new executable
ssh -i "$keylocation/aucbot-orch-prod.pem" "$PI_HOST" "sudo systemctl stop leader.service && \
mv /usr/local/bin/leader_new /usr/local/bin/leader && \
sudo systemctl start leader.service
" || die "Service restart failed."

echo "Deployment completed successfully."
