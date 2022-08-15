#!/bin/bash
echo "Starting node deployment."

die() {
  printf -- '%s\n' "$*"
  exit 1
}

if test -z "$PI_HOST"; then
  echo "Missing PI_HOST env variable, exiting."
  exit 1
fi

ssh "$PI_HOST" "mkdir -p /tmp/services && sudo mkdir -p /usr/local/bin/quillsecure" || die "Failed to create quillsecure directory"

# first copy config file
scp "$(pwd)/node/quillsecure_prod.yaml" "$PI_HOST":/tmp/quillsecure_node_prod.yaml || die "Failed to copy quillsecure_prod.yaml."
ssh "$PI_HOST" "sudo mv /tmp/quillsecure_node_prod.yaml /usr/local/bin/quillsecure/quillsecure_node.yaml" || die "Failed to copy quillsecure_prod.yaml."

# then copy binary
scp "$(pwd)/bin/node" "$PI_HOST":/tmp/node_new || die "Failed to copy node_new."

# then restart service, swapping in the new executable
scp "$(pwd)/deploy/assets/node.service" "$PI_HOST":/tmp/services/node.service || die "Failed to copy node service."

ssh "$PI_HOST" "sudo mv /tmp/services/node.service /etc/systemd/system/node.service
sudo systemctl daemon-reload && \
(sudo systemctl stop node.service || true) && \
sudo mv /tmp/node_new /usr/local/bin/quillsecure/node && \
sudo chmod 777 /usr/local/bin/quillsecure/node
sudo systemctl start node.service
" || die "Service restart failed."

#ssh "$PI_HOST" "sudo mv /tmp/node_new /usr/local/bin/quillsecure/node" || die "TODO mv failed"

echo "Deployment completed successfully."
