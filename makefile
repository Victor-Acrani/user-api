# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# up:
# 	go run app/user-service/main.go

up:
	go run app/user-service/main.go | go run app/tooling/logfmt/main.go	