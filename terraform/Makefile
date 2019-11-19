TERRAFORM_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
TERRAFORM_WORKSPACES := $(shell cd $(TERRAFORM_DIR) && terraform workspace list)

ensure_workspace_exists:
	cd ${TERRAFORM_DIR} && \
	([[ ! "$(TERRAFORM_WORKSPACES)" =~ "$(environment)" ]] && terraform workspace new $(environment)) 2>&1 >/dev/null || true

switch_workspace: ensure_workspace_exists
	cd ${TERRAFORM_DIR} && \
	terraform workspace select $(environment)

init:
	cd ${TERRAFORM_DIR} && \
	terraform init \
	-backend=true \
	-backend-config="role_arn=${role_arn}" \
	-backend-config="bucket=${bucket}" \
	-backend-config="key=tasque-test-interruption/terraform.tfstate" \
	-backend-config="region=${region}"

plan: switch_workspace
	cd ${TERRAFORM_DIR} && \
	terraform plan \
	-var "role_arn=${role_arn}"

apply: switch_workspace
	cd ${TERRAFORM_DIR} && \
	terraform apply \
	-auto-approve \
	-var "role_arn=${role_arn}"

destroy: switch_workspace
	cd ${TERRAFORM_DIR} && \
	terraform destroy \
	-var "role_arn=${role_arn}"
