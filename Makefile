dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-dev" |.stages.staging |= (.domain = "invite.dev.unee-t.com" | .zone = "dev.unee-t.com")' up.json.in > up.json
	up deploy production

demo:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-demo" |.stages.staging |= (.domain = "invite.demo.unee-t.com" | .zone = "demo.unee-t.com")' up.json.in > up.json
	up deploy production

prod:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-prod" |.stages.staging |= (.domain = "invite.unee-t.com" | .zone = "unee-t.com")' up.json.in > up.json
	up deploy production


.PHONY: dev demo prod
