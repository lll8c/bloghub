.PHONY: docker
docker:
	@del webook || true
	@set GOOS=linux
	@set GOARCH=amd64
	@go build -o webook .
	@set GOOS=windows
	@docker rmi -f flycash/webook:v0.0.1
	@docker build -t flycash/webook:v0.0.1 .
