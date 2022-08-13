#!/bin/bash
echo "Starting leader deployment."

die() {
  printf -- '%s\n' "$*"
  exit 1
}

if test -z "$PI_HOST"; then
  echo "Missing PI_HOST env variable, exiting."
  exit 1
fi

ssh "$PI_HOST" "sudo mkdir -p /usr/local/bin/quillsecure" || die "Failed to create quillsecure directory"

# first copy config file
scp "$(pwd)/leader/quillsecure_prod.yaml" "$PI_HOST":/tmp/quillsecure_leader_prod.yaml || die "Failed to copy quillsecure_prod.yaml."
ssh "$PI_HOST" "sudo mv /tmp/quillsecure_leader_prod.yaml /usr/local/bin/quillsecure/quillsecure_leader.yaml" || die "Failed to copy quillsecure_prod.yaml."

# then copy binary
scp "$(pwd)/bin/leader" "$PI_HOST":/tmp/leader_new || die "Failed to copy leader_new."

# TODO service
# then restart service, swapping in the new executable
#ssh "$PI_HOST" "sudo systemctl stop leader.service && \
#mv /tmp/leader_new /usr/local/bin/quillsecure/leader && \
#sudo systemctl start leader.service
#" || die "Service restart failed."

ssh "$PI_HOST" "sudo mv /tmp/leader_new /usr/local/bin/quillsecure/leader" || die "TODO mv failed"

echo "Deployment completed successfully."
