git_sha = $(shell git rev-parse --short HEAD)
git_branch = $(shell git rev-parse --abbrev-ref HEAD)
git_summary = $(shell git describe --tags --dirty --always)
build_date = $(shell date '+%y-%m-%d %H:%M)
version = $(shell cat VERSION)
arch ?= amd64
docker_repo = 770136283015.dkr.ecr.us-west-2.amazonaws.com/tasque-go
docker_tag ?= latest
region = us-west-2
registry_id = 770136283015
docker-login:
	$(eval loginstring = $(shell aws ecr get-login --region ${region} --no-include-email --registry-ids ${registry_id} | sed 's|https://||'))
	$(eval aws_ecs_repo_domain = $(subst https://,,$(lastword $(loginstring))))
	@echo "Logging in"
	@$(loginstring)

build:
	docker build \
		--no-cache \
		--build-arg arch=${arch} \
		--build-arg version=${version} \
		--build-arg git_summary=${git_summary} \
		--build-arg build_date=${build_date} \
		--build-arg git_sha=${git_sha} \
		--build-arg git_branch=${git_branch} \
		-t tasque/tasque:${arch} .
	docker tag tasque/tasque:${arch} ${docker_repo}:${docker_tag}

push: docker-login
	docker push ${docker_repo}:${docker_tag}

include ./terraform/Makefile