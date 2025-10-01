login:
	docker login -u kamathakki

push-redis:
	docker tag miko-stock-automation-redis:latest kamathakki/miko-stock-automation-redis:latest
	docker push kamathakki/miko-stock-automation-redis:latest

push-postgres:
	docker tag miko-stock-automation-postgres:latest kamathakki/miko-stock-automation-postgres:latest
	docker push kamathakki/miko-stock-automation-postgres:latest

push-backend-iam:
	docker tag miko-stock-automation-backend-iam:latest kamathakki/miko-stock-automation-backend-iam:latest
	docker push kamathakki/miko-stock-automation-backend-iam:latest

push-backend-stockkeepingunit:
	docker tag miko-stock-automation-backend-stockkeepingunit:latest kamathakki/miko-stock-automation-backend-stockkeepingunit:latest
	docker push kamathakki/miko-stock-automation-backend-stockkeepingunit:latest

push-backend-warehouse:
	docker tag miko-stock-automation-backend-warehouse:latest kamathakki/miko-stock-automation-backend-warehouse:latest
	docker push kamathakki/miko-stock-automation-backend-warehouse:latest

push-backend-gateway:
	docker tag miko-stock-automation-backend-gateway:latest kamathakki/miko-stock-automation-backend-gateway:latest
	docker push kamathakki/miko-stock-automation-backend-gateway:latest

push-all:
	make push-redis
	make push-postgres
	make push-backend-iam
	make push-backend-stockkeepingunit
	make push-backend-warehouse
	make push-backend-gateway
    