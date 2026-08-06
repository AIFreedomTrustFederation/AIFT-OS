.RECIPEPREFIX := >
.PHONY: help build doctor status verify verify-fast format smoke coverage architecture install-hooks test registry dashboard deps plugins safe-sync inspect

help:
>bash scripts/dev.sh help

build:
>bash scripts/dev.sh build

verify:
>bash scripts/dev.sh verify

verify-fast:
>bash scripts/dev.sh verify-fast

format:
>bash scripts/dev.sh format

smoke:
>bash scripts/dev.sh smoke

coverage:
>bash scripts/dev.sh coverage

architecture:
>bash scripts/dev.sh architecture

install-hooks:
>bash scripts/dev.sh install-hooks

doctor:
>./aift-os.sh doctor

status:
>./aift-os.sh status

test:
>bash scripts/dev.sh test

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
