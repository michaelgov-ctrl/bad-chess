#!/bin/bash
set -eu

TIMEZONE=America/New_York
USERNAME=bad-chess

export LC_ALL=en_US.UTF-8

apt update

timedatectl set-timezone ${TIMEZONE}

# uncomment 'en_US.UTF-8 UTF-8' from the locale gen config & then unpack it
sed -i '/en_US.UTF-8 UTF-8/ s/^#//' /etc/locale.gen
locale-gen

# create local user & strip root for service
useradd --create-home --shell "/bin/bash" --group sudo "${USERNAME}"
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"
cp -a /root/.ssh /home/${USERNAME} && chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}/.ssh

# basic firewall access
apt --yes install ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# get that sweet sweet fail2ban
apt --yes install fail2ban

# install & configure caddy
# https://caddyserver.com/docs/install#debian-ubuntu-raspbian
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl libnss3-tools
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

sudo mv -f ~/remote/Caddyfile /etc/caddy/
sudo systemctl reload caddy

# upgrade packages & replace config files if available
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# enable bad-chess service
sudo mv ~/remote/bad-chess-server /home/bad-chess/
sudo mv ~/remote/bad-chess.service /etc/systemd/system/
sudo systemctl enable bad-chess
sudo systemctl restart bad-chess

reboot
# tired