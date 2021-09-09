#!/bin/bash
set -eu

# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

# Set the timezone for the server. A full list of available timezones can be found by
# running timedatectl list-timezones.
# TIMEZONE=Asia/Singapore

# Set the name of the new user to create.
# USERNAME=skel

# Prompt to enter a password for the PostgreSQL skel user (rather than hard-coding
# a password in this script).
read -p 'Enter password for skel DB user: ' DB_PASSWORD

# Force all output to be presented in en_US for the duration of this script. This avoids
# any 'setting locale failed' errors while this script is running, before we have
# installed support for all locales. Do not change this setting!
# export LC_ALL=en_US.UTF-8

# ==================================================================================== #
# SCRIPT LOGIC
# ==================================================================================== #

# Enable the 'universe' repository.
# add-apt-repository --yes universe

# Update all software packages. Using the --force-confnew flag means that configuration
# files will be replaced if newer ones are available.
# sudo apt update
# apt --yes -o Dpkg::Options::='--force-confnew' upgrade

# Set the system timezone and install all locales.
# timedatectl set-timezone ${TIMEZONE}
# apt --yes install locales-all

# Add the new user (and give them sudo privileges).
# useradd --create-home --shell '/bin/bash' --groups sudo '${USERNAME}'

# Force a password to be set for the new user the first time they log in.
# passwd --delete '${USERNAME}'
# chage --lastday 0 '${USERNAME}'

# Copy the SSH keys from the root user to the new user.
# rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# Configure the firewall to allow SSH, HTTP and HTTPS traffic.
# ufw allow 22
# ufw allow 80/tcp
# ufw allow 443/tcp
# ufw --force enable

# Install fail2ban.
# apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL.
# sudo apt --yes install postgresql

# Start PostgreSQL cluster
# sudo pg_ctlcluster 12 main start

# Set up the skel DB and create a user account with the password entered earlier.
psql -c "CREATE DATABASE skel"
psql -d skel -c "CREATE EXTENSION IF NOT EXISTS citext"
psql -d skel -c "CREATE ROLE skel WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# Add a DSN for connecting to the skel database to the system-wide environment
# variables in the /etc/environment file.
echo "SKEL_DB_DSN=postgres://skel:${DB_PASSWORD}@localhost/skel" | sudo tee -a /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
sudo apt --yes install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/gpg.key" | sudo tee /etc/apt/trusted.gpg.d/caddy-stable.asc
curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt" | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

echo "Script complete!"
# echo "Rebooting..."
# reboot