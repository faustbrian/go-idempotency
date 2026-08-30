SHELL := /usr/bin/env bash

GOLIB ?= golib

include verification/package.mk

.PHONY: check ci inventory repository-check

check:
	$(GOLIB) check --all

ci: repository-check check

inventory repository-check:
	$(GOLIB) repository check
