SHELL := /bin/bash

AZURE_HARNESS := ./tests/azure_test_harness
DOCKER_HARNESS := ./tests/docker_test_harness

build:
	go build -v .

test:
	go test -v ./internal/...

accTest:
	if [ -f local.env ]; then source ./local.env; fi && TF_ACC=1 go test -count 1 -timeout 60m -v ./internal/...

setupAccTest:
	cd $(MAKE_PATH); terraform init -reconfigure && terraform apply -auto-approve

destroyAccTest:
	cd $(MAKE_PATH) && terraform destroy -auto-approve

accTestLifecycle:
	$(MAKE) accTest; TEST_CODE=$$?; $(MAKE) $(MAKE_DESTROY_TARGET) && exit $$TEST_CODE

# AZURE
azureSetupAccTest:
	MAKE_PATH=$(AZURE_HARNESS) $(MAKE) setupAccTest

azureDestroyAccTest:
	cd $(AZURE_HARNESS) && terraform init -reconfigure && (terraform state rm azurerm_mssql_elasticpool.this; terraform destroy -auto-approve)

azureAccTest: azureSetupAccTest
	MAKE_DESTROY_TARGET=azureDestroyAccTest $(MAKE) accTestLifecycle


#DOCKER
dockerSetupAccTest:
	MAKE_PATH=$(DOCKER_HARNESS) $(MAKE) setupAccTest

dockerDestroyAccTest:
	MAKE_PATH=$(DOCKER_HARNESS) $(MAKE) destroyAccTest

dockerAccTest: dockerSetupAccTest
	MAKE_DESTROY_TARGET=dockerDestroyAccTest $(MAKE) accTestLifecycle
