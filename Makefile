REGION:=ap-southeast-1
PROFILE := ins-dev
define ssm
$(shell aws --profile $(PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

DEVUPJSON = '.profile |= "$(PROFILE)" \
		  | .vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .role |= "arn:aws:iam::$(call ssm,ACCOUNT_ID):role/invitefromqueue_lambda_function" \
		  | .vpc.securityGroups |= [ "$(call ssm,DEFAULT_SECURITY_GROUP)" ]'

dev:
	go generate
	jq $(DEVUPJSON) project.dev.json.in > project.dev.json
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev deploy

devlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev logs
