dev:
	jq '.profile |= "uneet-dev" |.stages.staging |= (.domain = "processinvites.dev.unee-t.com" | .zone = "dev.unee-t.com")' up.json.in > up.json
	up

demo:
	jq '.profile |= "uneet-demo" |.stages.staging |= (.domain = "processinvites.demo.unee-t.com" | .zone = "demo.unee-t.com")' up.json.in > up.json
	up

prod:
	jq '.profile |= "uneet-prod" |.stages.staging |= (.domain = "processinvites.unee-t.com" | .zone = "unee-t.com")' up.json.in > up.json
	up


.PHONY: dev demo prod
