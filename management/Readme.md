-- Bank is in config file as the only entity permitted to change identity permissions
-- For a new client: the bank sends transaction to network identity family to list new client as permitted transactor for transaction family specific to the client's account and identity family

-- Multiple roles: admin (bank), rule setters, transactors
-- Admin sets rule setters and Rule Setters set transactors?

-- Default groups: 
    -- Suppliers with rule (for example 2 sigs). How to make all suppliers part of default group?
    -- Default group for initiators?

-- Paths configuration for deployment:
    -- If we want to split data across many blockchains, then we have 2 options:
        -- create multiple config dirs and run sawtooth-validator with --config-dir pointing to account dependent dir
        -- use docker images with one configuration dir for all runs but mount account dependent dir to docker image
    -- Regardless of option, genesis block will (probably) have to be moved to the right dir

-- NOTE: need to stress test multiple blockchain scenario


Log of sys admin (to form the basis of future scripts):
======================================================

# installing sawtooth on ubuntu
$ sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 8AA7AF1F1091A5FD
###### the next one should be dependent on the version
$ sudo add-apt-repository 'deb http://repo.sawtooth.me/ubuntu/1.0/stable xenial universe'
$ sudo apt-get update
$ sudo apt-get install -y sawtooth

# On Linux, the default path settings are:
#
#   conf_dir = "/etc/sawtooth"
#   key_dir  = "/etc/sawtooth/keys"
#   data_dir = "/var/lib/sawtooth"
#   log_dir  = "/var/log/sawtooth"
#   policy_dir  = "/etc/sawtooth/policy"

# keys for new user (default current $USER):
$ sawtooth keygen new_user

# Create keys for "bank" (keys were stored in /home/majed/.sawtooth/keys/bank.*). If I had created a new user, bank, than I suspect the keys would have been stored under ~bank. bank is the entity that I want to give the exclusive access to Identity TP:
$ sawtooth keygen bank


# validator needs own keys to sign blocks
majed@pluto:~$ sudo sawadm keygen # can't start validator without doing this first
# output [sudo] password for majed:
# output writing file: /etc/sawtooth/keys/validator.priv
# output writing file: /etc/sawtooth/keys/validator.pub



# Genesis block (all data goes into /var/lib/sawtooth)
# to restart from scratch, delete all files in /var/lib/sawtooth and redo this
$ sawset genesis
$ sudo -u sawtooth sawadm genesis config-genesis.batch

# running validator (note that /etc/sawtooth already had the same example file; i didn't need to copy it from $GOPATH/...)
$ cp sawtooth-core/validator/packaging/validator.toml.example /etc/sawtooth


# (this is for Windows Subsystem for Linux and NOT Ubuntu?) working on lmdb: copied package and modified line 57 of lmdb_nolock_database.py, writemap=True, to writemap=False
$ cp -r /usr/lib/python3/dist-packages/sawtooth_validator /mnt/d/code/python/modules
