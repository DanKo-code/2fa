gen_docs:
	swag fmt;
	swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./docs;
