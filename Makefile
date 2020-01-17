REGION := ap-southeast-1
PROFILE := ins-dev

define ssm
$(shell aws --profile $(PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

DEVUPJSON = '.profile |= "$(PROFILE)" \
		  | .vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .role |= "arn:aws:iam::$(call ssm,ACCOUNT_ID):role/invitefromqueue_lambda_function" \
		  | .vpc.securityGroups |= [ "$(call ssm,DEFAULT_SECURITY_GROUP)" ]'

test:
	@echo $(PROFILE) at $(REGION)

dev:
	go generate
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev deploy

demo:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env demo deploy

demologs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env demo logs -f

prod:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env prod deploy

devlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev logs -f

prodlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env prod logs -f
