.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Welcome to file_to_text_example. Please use \`make <target>\` where <target> is one of"
	@echo " "
	@echo "  Next commands are only for dev environment with nextcloud-docker-dev!"
	@echo "  They should run from the host you are developing on(with activated venv) and not in the container with Nextcloud!"
	@echo "  "
	@echo "  build-push        build image and upload to ghcr.io"
	@echo "  "
	@echo "  run               install file_to_text_example for Nextcloud Last"
	@echo "  run27             install file_to_text_example for Nextcloud 27"
	@echo "  "
	@echo "  For development of this example use GoLand run configurations. Development is always set for last Nextcloud."
	@echo "  First run 'file_to_text_example' and then 'make registerX', after that you can use/debug/develop it and easy test."
	@echo "  "
	@echo "  register          perform registration of running 'file_to_text_example' into 'manual_install' deploy daemon."
	@echo "  register27        perform registration of running 'file_to_text_example' into 'manual_install' deploy daemon."

.PHONY: build-push
build-push:
	docker login ghcr.io
	docker buildx build --push --platform linux/arm64/v8,linux/amd64 --tag ghcr.io/cloud-py-api/file_to_text_example:1.3.0 --tag ghcr.io/cloud-py-api/file_to_text_example:latest .

.PHONY: run
run:
	docker exec master-nextcloud-1 sudo -u www-data php occ app_api:app:unregister file_to_text_example --silent --force || true
	docker exec master-nextcloud-1 sudo -u www-data php occ app_api:app:register file_to_text_example \
		--force-scopes \
		--info-xml https://raw.githubusercontent.com/cloud-py-api/file_to_text_example/main/appinfo/info.xml

.PHONY: run27
run27:
	docker exec master-stable27-1 sudo -u www-data php occ app_api:app:unregister file_to_text_example --silent --force || true
	docker exec master-stable27-1 sudo -u www-data php occ app_api:app:register file_to_text_example \
		--force-scopes \
		--info-xml https://raw.githubusercontent.com/cloud-py-api/file_to_text_example/main/appinfo/info.xml

.PHONY: register
register:
	docker exec master-nextcloud-1 sudo -u www-data php occ app_api:app:unregister file_to_text_example --silent --force || true
	docker exec master-nextcloud-1 sudo -u www-data php occ app_api:app:register file_to_text_example manual_install --json-info \
  "{\"id\":\"file_to_text_example\",\"name\":\"file_to_text_example\",\"daemon_config_name\":\"manual_install\",\"version\":\"1.0.0\",\"secret\":\"12345\",\"port\":10070,\"scopes\":[\"FILES\", \"NOTIFICATIONS\"],\"system\":0}" \
  --force-scopes --wait-finish

.PHONY: register27
register27:
	docker exec master-stable27-1 sudo -u www-data php occ app_api:app:unregister file_to_text_example --silent --force || true
	docker exec master-stable27-1 sudo -u www-data php occ app_api:app:register file_to_text_example manual_install --json-info \
  "{\"id\":\"file_to_text_example\",\"name\":\"file_to_text_example\",\"daemon_config_name\":\"manual_install\",\"version\":\"1.0.0\",\"secret\":\"12345\",\"port\":10070,\"scopes\":[\"FILES\", \"NOTIFICATIONS\"],\"system\":0}" \
  --force-scopes --wait-finish
