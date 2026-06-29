.RECIPEPREFIX := >
.PHONY: help build doctor status verify test registry dashboard deps plugins safe-sync inspect

help:
>./aift-os.sh help

build:
>sh install/01-build.sh

doctor:
>./aift-os.sh doctor

status:
>./aift-os.sh status

verify:
>./aift-os.sh verify

test:
>sh tests/go-smoke.sh

registry:
>./aift-os.sh registry

dashboard:
>./aift-os.sh dashboard

deps:
>./aift-os.sh deps

plugins:
>./aift-os.sh plugins

safe-sync:
>./aift-os.sh sync --safe

inspect:
>sh scripts/inspect.sh
