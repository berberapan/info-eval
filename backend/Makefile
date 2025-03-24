include .envrc

.PHONY: help	
## help: prints this message
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n "Are you sure [y/n] " && read ans && [ $${ans: -N} = y ]
