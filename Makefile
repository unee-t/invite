REGION:=ap-southeast-1

devlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev logs -f

dev:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env dev deploy

demo:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env demo deploy

prod:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(REGION) --env prod deploy
