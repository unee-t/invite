curl -i \
	-H "Authorization: Bearer $(aws --profile uneet-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" \
	http://localhost:3000/

curl -i -X POST \
	-H "Content-Type: application/json" --data @invites.json \
	-H "Authorization: Bearer $(aws --profile uneet-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" \
	http://localhost:3000/
