default: help

help: ## Show help
	@grep -E '(^[0-9a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-25s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

install-deps: ## Install dependencies
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Hack to run create-migration target with arguments
ifeq (create-migration,$(firstword $(MAKECMDGOALS)))
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(RUN_ARGS):;@:)
endif

create-migration:  ## Create migration file. Example: make create-migration create_users_table s
	@goose -dir migrations -s create $(RUN_ARGS)