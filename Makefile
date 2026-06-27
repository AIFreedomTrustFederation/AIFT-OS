.PHONY: help doctor status verify dashboard deps plugins safe-sync

help:
@./aift-os.sh help

doctor:
@./aift-os.sh doctor

status:
@./aift-os.sh status

verify:
@./aift-os.sh verify

dashboard:
@./aift-os.sh dashboard

deps:
@./aift-os.sh deps

plugins:
@./aift-os.sh plugins

safe-sync:
@./aift-os.sh sync --safe
