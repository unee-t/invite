dev:
	go generate
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env dev deploy

devlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env dev logs -f

demo:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env demo deploy

demologs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env demo logs -f

prod:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env prod deploy

prodlogs:
	@echo $$AWS_ACCESS_KEY_ID
	apex -r $(AWS_REGION) --env prod logs -f
