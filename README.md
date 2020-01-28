[Process Invitations in the BZ database](https://github.com/unee-t/bugzilla-customisation/wiki/Process-Invitations-in-the-BZ-database)

On [Docker Hub](https://hub.docker.com/r/uneet/invite) for [local development](https://github.com/unee-t/bugzilla-customisation/blob/master/docker-compose.yml)

[![Build Status](https://travis-ci.org/unee-t/invite.svg?branch=master)](https://travis-ci.org/unee-t/invite)

# Different environments

##################################
# This needs to be reviewed <-- these are hardcoded values that are environment specific.
##################################

This is codified over at: https://github.com/unee-t/env

## dev environment

* AWS account id: 812644853088
* profile: uneet-dev
* https://invite.dev.unee-t.com

## demo environment

* AWS account id: 915001051872
* profile: uneet-demo
* https://invite.demo.unee-t.com

## prod environment

* AWS account id: 192458993663
* profile: uneet-prod
* https://invite.unee-t.com

##################################
# END This needs to be reviewed
##################################

# Test plan

`/health_check` for a sanity check with the database connection
