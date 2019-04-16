git_sha = $(shell git rev-parse --short HEAD)
git_branch = $(shell git rev-parse --abbrev-ref HEAD)
git_summary = $(shell git describe --tags --dirty --always)
build_date = $(shell date)
version = $(shell cat VERSION)
arch ?= amd64
docker_repo = 770136283015.dkr.ecr.us-west-2.amazonaws.com/tasque-go
region = us-west-2
registry_id = 770136283015
docker-login:
	$(eval loginstring = $(shell aws ecr get-login --region ${region} --no-include-email --registry-ids ${registry_id} | sed 's|https://||'))
	$(eval aws_ecs_repo_domain = $(subst https://,,$(lastword $(loginstring))))
	@echo "Logging in"
	@$(loginstring)

build: docker-login
	go get
	CGO_ENABLED=0 GOOS=linux GOARCH=${arch} go build \
	-a -installsuffix cgo \
	-ldflags "-X 'main.Version=${version}' -X 'main.GitSummary=${git_summary}' -X 'main.BuildDate=${build_date}' -X main.GitCommit=${git_sha} -X main.GitBranch=${git_branch}" \
	-o tasque .
	docker build -t tasque/tasque:${arch} .
	docker tag tasque/tasque:${arch} ${docker_repo}

push: docker-login
	docker push ${docker_repo}
