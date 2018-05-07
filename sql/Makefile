refresh:
	curl -O https://raw.githubusercontent.com/unee-t/bz-database/master/scripts/process_one_invitation_all_scenario_v3.0.sql

test:
	mysql -h auroradb.dev.unee-t.com -P 3306 -u bugzilla --password=$(aws --profile uneet-dev ssm get-parameters --names MYSQL_PASSWORD --with-decryption --query Parameters[0].Value --output text) bugzilla < script.sql
