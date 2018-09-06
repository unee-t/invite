DEVUPJSON = '.profile |= "uneet-dev" \
		  |.stages.production |= (.domain = "unit.dev.unee-t.com" | .zone = "dev.unee-t.com") \
		  | .actions[0].emails |= ["kai.hendry+unitdev@unee-t.com"] \
		  | .lambda.vpc.subnets |= [ "subnet-0e123bd457c082cff", "subnet-0ff046ccc4e3b6281", "subnet-0e123bd457c082cff" ] \
		  | .profile |= "uneet-dev" \
		  | .lambda.vpc.security_groups |= [ "sg-0b83472a34bc17400", "sg-0f4dadb564041855b" ]'

dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(DEVUPJSON) up.json.in > up.json
	up deploy production

demo:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-demo" |.stages.production |= (.domain = "invite.demo.unee-t.com" | .zone = "demo.unee-t.com") | .actions[0].emails |= ["kai.hendry+invitedemo@unee-t.com"]' up.json.in > up.json
	up deploy production

prod:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-prod" |.stages.production |= (.domain = "invite.unee-t.com" | .zone = "unee-t.com")| .actions[0].emails |= ["kai.hendry+inviteprod@unee-t.com"]' up.json.in > up.json
	up deploy production

testdev:
	curl -H "Authorization: Bearer $(shell aws --profile uneet-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://invite.dev.unee-t.com/version

testdemo:
	curl -H "Authorization: Bearer $(shell aws --profile uneet-demo ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://invite.demo.unee-t.com/version

testprod:
	curl -H "Authorization: Bearer $(shell aws --profile uneet-prod ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://invite.unee-t.com/version

.PHONY: dev demo prod
