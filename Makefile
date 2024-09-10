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