# New Pi Setup

## Create user
`sudo useradd quillsecure`
`sudo apt-get install sqlite`

## Create service


## Leader
`sudo apt-get install gcc`

`sudo apt-get install sqlite3`

`sudo apt-get install unzip`

`sudo apt-get install git`

https://www.digitalocean.com/community/tutorials/how-to-install-go-on-debian-10
To install go

## NGINX setup
https://www.linode.com/docs/guides/how-to-install-and-use-nginx-on-ubuntu-20-04/
`sudo unlink /etc/nginx/sites-enabled/default`

### Example commands
`systemctl status leader.service`
Tail logs: `sudo journalctl -u leader.service -f`
alias tail='sudo journalctl -u leader.service -f'

