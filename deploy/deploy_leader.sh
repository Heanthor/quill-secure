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

ssh "$LINODE_HOST" "sudo mkdir -p /usr/local/bin/quillsecure && \
rm -r /tmp/src && \
sudo mkdir -p /tmp/src && \
sudo mkdir -p /var/www/quillsecure.com && \
sudo mkdir -p /etc/caddy && \
sudo chown quillsecure /usr/local/bin/quillsecure" || die "Failed to create directories"

# first copy config file
scp "$(pwd)/leader/quillsecure_prod.yaml" "$LINODE_HOST":/tmp/quillsecure_leader_prod.yaml || die "Failed to copy quillsecure_prod.yaml."
ssh "$LINODE_HOST" "sudo mv /tmp/quillsecure_leader_prod.yaml /usr/local/bin/quillsecure/quillsecure_leader.yaml" || die "Failed to copy quillsecure_prod.yaml."

# then copy sources and compile binary
rm $(pwd)/bin/leader_src.zip
zip -r bin/leader_src.zip $(git ls-files)
scp "$(pwd)/bin/leader_src.zip" "$LINODE_HOST":/tmp/src/leader_src.zip || die "Failed to copy leader sources."
ssh "$LINODE_HOST" "cd /tmp/src && \
unzip leader_src.zip && \
$remote_go mod download -x && \
env GOOS=linux $remote_go build -ldflags=\"-s -w\" -o /tmp/leader_new \$(ls -1 leader/*.go | grep -v _test.go) && \
rm /tmp/src/leader_src.zip
" || die "Failed to build sources"

# then restart service, swapping in the new executable
scp "$(pwd)/deploy/assets/leader.service" "$LINODE_HOST":/etc/systemd/system/leader.service || die "Failed to copy leader service."

ssh "$LINODE_HOST" "sudo systemctl daemon-reload && \
(sudo systemctl stop leader.service || true) && \
mv /tmp/leader_new /usr/local/bin/quillsecure/leader && \
sudo chmod 777 /usr/local/bin/quillsecure/leader && \
sudo systemctl start leader.service
" || die "Service restart failed."

# configure caddy
caddy fmt -overwrite "$(pwd)/deploy/assets/Caddyfile"
scp "$(pwd)/deploy/assets/Caddyfile" "$LINODE_HOST":/etc/caddy/sites/quillsecure.caddy || die "Failed to upload Caddyfile"

ssh "$LINODE_HOST" "cd /etc/caddy && \
sudo caddy validate
sudo caddy reload
" || die "Caddy configuration failed."

# configure static files
# TODO this should be moved into another executable!
scp -r "$(pwd)/../quillsecure-site/build" "$LINODE_HOST":/tmp || die "Failed to upload site."

ssh "$LINODE_HOST" "sudo rm -r /var/www/quillsecure.com && \
sudo mkdir -p /var/www/quillsecure.com && \
sudo mv /tmp/build/* /var/www/quillsecure.com" || die "Failed to move site files."

echo "Deployment completed successfully."
