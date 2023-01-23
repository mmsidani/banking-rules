#!/bin/bash

# Install the sawtooth package

sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 8AA7AF1F1091A5FD
# ACTHUNG this next one depends on the version. Fix to move version number to a config file
sudo add-apt-repository 'deb http://repo.sawtooth.me/ubuntu/1.0/stable xenial universe'
sudo apt-get update
sudo apt-get install -y sawtooth

# Deploy configuration files. 
# Important notes:
#   -- the previous step (i.e., installation) puts example files in /etc/sawtooth
#   -- we assume here that we have one blockchain deployment. to be modified if we split across many. see ReadMe.txt for options

# TODO TODO TODO TODO TODO the following assumes 1 blockchain. to be modified if/when we split

sudo cp validator.toml settings.toml rest_api.toml path.toml log_config.toml cli.toml /etc/sawtooth

# Generate keys for the bank (needed to validate blocks) and create genesis block. 

# TODO these are the instructions for developers -- not sysadmin . Eventually we want sysadmin
# TODO assumes 1 blockchain 

# create keys for the bank. output goes to ~/.sawtooth/keys/bank.{priv,pub} (whatever ~ happens to be)
sawtooth keygen bank
# create genesis block. TODO this next command creates a file 'config-genesis.batch' in the current dir. Must make sure we have writing privileges in current dir.
sudo sawset genesis
sudo -u sawtooth sawadm genesis config-genesis.batch
# now clean up
sudo rm -f config-genesis.batch

# this creates keys for the validator. output to /etc/sawtooth/keys/validator.{priv, pub}
sudo sawadm keygen

# Set identity transaction family permissions

# TODO TODO TODO shouldn't really belong in my home dir
bank_key=`cat /home/majed/.sawtooth/keys/bank.pub`

# only bank can change identity settings
sawset proposal create sawtooth.identity.allowed_keys=${bank_key}
