#!/bin/bash

# nohup so processes don't die if terminal is killed. -b for background. TODO do something about output log, maybe add today's date to the name
sudo -b -u sawtooth nohup sawtooth-validator -vv &> /var/tmp/sawtooth_validator.log < /dev/null
sudo -b -u sawtooth nohup sawtooth-rest-api -vv &> /var/tmp/sawtooth_restAPI.log < /dev/null

# settings must run the first time we fire everything up but is only required then. TODO this blocks. It seems to be needed to set the key that can modify settings. When I ran it I go the following message (the hex is my pub key ~/.sawtooth/keys/majed.pub):
# Setting setting sawtooth.settings.vote.authorized_keys changed from None to 030509c81f5d5e927cd2fbe17ac1e90866e53deec7bba53670afec5e562ef53f9d
sudo -u sawtooth settings-tp -vv