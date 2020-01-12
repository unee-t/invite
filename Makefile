REGION:=ap-southeast-1

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
