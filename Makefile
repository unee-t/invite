REGION:=ap-southeast-1

dev:
	go generate
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev deploy

devlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev logs -f
