#!/bin/bash
echo "Starting leader deployment."
remote_go=/usr/local/go/bin/go

die() {
  printf -- '%s\n' "$*"
  exit 1
}

if test -z "$LINODE_HOST"; then
  echo "Missing LINODE_HOST env variable, exiting."
  exit 1
fi

ssh "$LINODE_HOST" "sudo mkdir -p /usr/local/bin/quillsecure && rm -r /tmp/src && sudo mkdir -p /tmp/src && sudo chown quillsecure /usr/local/bin/quillsecure" || die "Failed to create directories"

# first copy config file
scp "$(pwd)/leader/quillsecure_prod.yaml" "$LINODE_HOST":/tmp/quillsecure_leader_prod.yaml || die "Failed to copy quillsecure_prod.yaml."
ssh "$LINODE_HOST" "sudo mv /tmp/quillsecure_leader_prod.yaml /usr/local/bin/quillsecure/quillsecure_leader.yaml" || die "Failed to copy quillsecure_prod.yaml."

# then copy sources and compile binary
zip -r bin/leader_src.zip $(git ls-tree --name-only -r HEAD)
scp "$(pwd)/bin/leader_src.zip" "$LINODE_HOST":/tmp/src/leader_src.zip || die "Failed to copy leader sources."
ssh "$LINODE_HOST" "cd /tmp/src && \
unzip leader_src.zip && \
$remote_go mod download -x && \
env GOOS=linux $remote_go build -ldflags=\"-s -w\" -o /tmp/leader_new \$(ls -1 leader/*.go | grep -v _test.go) && \
rm /tmp/src/leader_src.zip
" || die "Failed to build sources"

# then restart service, swapping in the new executable
scp "$(pwd)/deploy/assets/leader.service" "$LINODE_HOST":/etc/systemd/system/leader.service || die "Failed to copy leader service."

ssh "$LINODE_HOST" "(sudo systemctl stop leader.service || true) && \
mv /tmp/leader_new /usr/local/bin/quillsecure/leader && \
sudo chmod 777 /usr/local/bin/quillsecure/leader && \
sudo systemctl daemon-reload && \
sudo systemctl start leader.service
" || die "Service restart failed."

echo "Deployment completed successfully."
