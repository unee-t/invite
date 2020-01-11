DEVUPJSON = '.profile |= "ins-dev" \
		  |.stages.production |= (.domain = "invite.dev.ins.unee-t.com" | .zone = "dev.ins.unee-t.com") \
		  | .actions[0].emails |= ["invite+dev.ins@unee-t.com"] \
		  | .lambda.vpc.subnets |= [ "subnet-07c9e20d6d4e35a64", "subnet-0f1579f2b82518454", "subnet-05f258f76ec0a2bd5" ] \
		  | .lambda.vpc.security_groups |= [ "sg-d8dde8a6" ]'

dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(DEVUPJSON) up.json.in > up.json
	up deploy production

testdev:
	curl -H "Authorization: Bearer $(shell aws --profile ins-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://invite.dev.ins.unee-t.com/metrics

