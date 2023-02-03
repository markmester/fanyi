#! Environment file to use when deploying - see .env.example
#! DO NOT change this variable; instead, to change the environment, run the Makefile as `ENV=config/.env Make <command>`
ENV 	?= config/.env.example
COMMIT 	:= $(shell git rev-parse HEAD)
VERSION := $(shell git describe --tags)-${COMMIT}


include $(ENV)
export

.PHONY: all
all: help

.PHONY: VERSION
version:
	@echo $(VERSION)

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Lint the project
	go fmt ./pkg/... 
	go vet ./pkg/... 
	staticcheck ./pkg/...

.PHONY: build # Build app services (docker)
build: build-docker

.PHONY: deploy ## Deploy to ECS
deploy: build push ecs-generate ecs-deploy

.PHONY: up ## Run local contianer
up: 
	docker run -d --rm --name fanyi $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/fanyi-slackbot:$(VERSION)

.PHONY: down ## Stop local contianer
down: 
	docker rm -f fanyi

.PHONY: build-docker
build-docker: lint ## Build app services (docker)
	@printf "\033[36m==> %s\033[0m\n" "Building services (Docker)..."
	@docker-compose build --force-rm --no-cache

.PHONY: push
push: ## Push services to ECR
	@printf "\033[36m==> %s\033[0m\n" "Pushing services to container registry..."
	@aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ECR)
	@docker-compose push


.PHONY: ecs-generate
ecs-generate: config ## Generate ECS service spec (remote)	
	@printf "\033[36m==> %s\033[0m\n" "Generating service spec (ECS)..."
	@aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ECR)
	@docker context create ecs fanyi-ecs-context --from-env 2>/dev/null; true
	@docker --context fanyi-ecs-context compose --project-name fanyi-slackbot convert > /tmp/fanyi_cf.yml
	
.PHONY: ecs-deploy
ecs-deploy: check ## Deploy ECS service specification (Cloudformation Stack) (remote)
	@printf "\033[36m==> %s\033[0m\n" "Deploying service spec (ECS)..."
	@aws cloudformation create-stack --template-body file:///tmp/fanyi_cf.yml --stack-name fanyi-slackbot --capabilities CAPABILITY_IAM
	@aws cloudformation wait stack-create-complete --stack-name fanyi-slackbot
	@printf "\033[36m==> %s\033[0m\n" "Deployment complete!"
