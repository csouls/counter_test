deploy:
	cd app; gcloud app deploy --quiet app.yaml
dev_appserver:
	cd app; dev_appserver.py --enable_watching_go_path=false app.yaml
