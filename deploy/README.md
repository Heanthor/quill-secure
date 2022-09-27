# New Pi Setup

## Create user
`sudo useradd quillsecure`
`sudo apt-get install sqlite`
`sudo systemctl enable node`

## Node
`libcamera-vid -t 0 --codec libav --libav-format mpegts -o "udp://192.168.0.26:9997"`

## Leader
`sudo apt-get install gcc`

`sudo apt-get install sqlite3`

`sudo apt-get install unzip`

`sudo apt-get install git`

`sudo systemctl enable leader`

https://www.digitalocean.com/community/tutorials/how-to-install-go-on-debian-10
To install go

## Caddy setup
https://caddyserver.com/docs/install#debian-ubuntu-raspbian

### Example commands
`systemctl status leader.service`
Tail logs: `sudo journalctl -u leader.service -f`
alias tail='sudo journalctl -u leader.service -f'

