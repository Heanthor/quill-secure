# New Pi Setup

## Create user
`sudo useradd quillsecure`
`sudo apt-get install sqlite`
`sudo systemctl enable node`

## Create service


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

