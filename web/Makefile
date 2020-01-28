# This script is called by the `deploy.sh` file in the /web folder (this folder)
# We use this to deploy the environments with Travis CI

# We create a function to simplify getting variables for aws parameter store.

define ssm
$(shell aws --profile $(TRAVIS_PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

# We prepare variables for up in UPJSON and PRODUPJSON.
# These variables are comming from AWS Parameter Store
# - STAGE
# - DOMAIN
# - EMAIL_FOR_NOTIFICATION_GENERIC
# - PRIVATE_SUBNET_1
# - PRIVATE_SUBNET_2
# - PRIVATE_SUBNET_3
# - DEFAULT_SECURITY_GROUP

UPJSON = '.profile |= "$(TRAVIS_PROFILE)" \
		  |.stages.production |= (.domain = "invite.$(call ssm,STAGE).$(call ssm,DOMAIN)" | .zone = "$(call ssm,STAGE).$(call ssm,DOMAIN)") \
		  | .actions[0].emails |= ["invite+$(call ssm,EMAIL_FOR_NOTIFICATION_GENERIC)"] \
		  | .lambda.vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .lambda.vpc.security_groups |= [ "$(call ssm,DEFAULT_SECURITY_GROUP)" ]'

#UPJSON for Production

PRODUPJSON = '.profile |= "$(TRAVIS_PROFILE)" \
		  |.stages.production |= (.domain = "invite.$(call ssm,DOMAIN)" | .zone = "$(call ssm,DOMAIN)") \
		  | .actions[0].emails |= ["invite+$(call ssm,EMAIL_FOR_NOTIFICATION_GENERIC)"] \
		  | .lambda.vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .lambda.vpc.security_groups |= [ "$(call ssm,DEFAULT_SECURITY_GROUP)" ]'

# We have everything, we can run up now.

dev:
	aws configure --profile $(TRAVIS_PROFILE) list
	jq $(UPJSON) up.json.in > up.json
	up deploy production

demo:
	aws configure --profile $(TRAVIS_PROFILE) list
	jq $(UPJSON) up.json.in > up.json
	up deploy production

prod:
	aws configure --profile $(TRAVIS_PROFILE) list
	jq $(PRODUPJSON) up.json.in > up.json
	up deploy production

testdev:
	curl -H "Authorization: Bearer $(shell aws --profile $(TRAVIS_PROFILE)-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://invite.$(call ssm,STAGE).$(call ssm,DOMAIN)/metrics

testdemo:	
	curl -H "Authorization: Bearer $(shell aws --profile $(TRAVIS_PROFILE) ssm get-parameters --names $(API_ACCESS_TOKEN) --with-decryption --query Parameters[0].Value --output text)" https://invite.$(call ssm,STAGE).$(call ssm,DOMAIN)/metrics	

testprod:	
	curl -H "Authorization: Bearer $(shell aws --profile $(TRAVIS_PROFILE) ssm get-parameters --names $(API_ACCESS_TOKEN) --with-decryption --query Parameters[0].Value --output text)" https://invite.$(call ssm,DOMAIN)/metrics	

testping:	
	curl -i -H "Authorization: Bearer $(shell aws --profile $(TRAVIS_PROFILE) ssm get-parameters --names $(API_ACCESS_TOKEN) --with-decryption --query Parameters[0].Value --output text)" https://invite.$(ZONE)/health_check 
