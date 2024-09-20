.PHONY: webookdocker webookk8s mysqlk8s
webookdocker:
	@del webook || true
	@set GOOS=linux
	@set GOARCH=arm
	@go build -tags=k8s -o webook .
	@set GOOS=windows
	@set GOARCH=amd64
	@docker rmi -f flycash/webook:v0.0.1
	@docker build -t flycash/webook:v0.0.1 .

webookk8s:
	@kubectl delete deployment webook
	@kubectl apply -f k8s-webook-deployment.yaml
	@kubectl apply -f k8s-webook-service.yaml

mysqlk8s:
	@kubectl delete deployment webook-mysql
	@kubectl apply -f k8s-mysql-deployment.yaml
	@kubectl apply -f k8s-mysql-pvc.yaml
	@kubectl apply -f k8s-mysql-pv.yaml
	@kubectl apply -f k8s-mysql-service.yaml

redisk8s:
	@kubectl delete deployment webook-redis
	@kubectl apply -f k8s-redis-deployment.yaml
	@kubectl apply -f k8s-redis-service.yaml

.PHONY: mock
mock:
	@mockgen -source=./internal/service/user.go -package=svcmocks -destination=./internal/service/mocks/user.mock.go
	@mockgen -source=./internal/service/code.go -package=svcmocks -destination=./internal/service/mocks/code.mock.go
	@mockgen -source=./internal/service/article.go -package=svcmocks -destination=./internal/service/mocks/article.mock.go
	@mockgen -source=./internal/repository/user.go -package=repomocks -destination=./internal/repository/mocks/user.mock.go
	@mockgen -source=./internal/repository/code.go -package=repomocks -destination=./internal/repository/mocks/code.mock.go
	@mockgen -source=./internal/repository/article/article_reader.go -package=repomocks -destination=./internal/repository/article/mocks/article_reader.mock.go
	@mockgen -source=./internal/repository/article/article_author.go -package=repomocks -destination=./internal/repository/article/mocks/article_author.mock.go
	@mockgen -source=./internal/repository/dao/user.go -package=daomocks -destination=./internal/repository/dao/mocks/user.mock.go
	@mockgen -source=./internal/repository/cache/user.go -package=cachemocks -destination=./internal/repository/cache/mocks/user.mock.go
	@mockgen -source=./pkg/ratelimit/types.go -package=limitmocks -destination=./pkg/ratelimit/mocks/ratelimit.mock.go
	@mockgen -package=redismocks -destination=./internal/repository/cache/redismocks/cmd.mock.go github.com/redis/go-redis/v9 Cmdable