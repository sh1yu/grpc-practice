TAG = 3

.PHONY: all
all: server push upgrade

.PHONY: server
server:
	@docker build -t psycore/grpc-practice-server:$(TAG) -f server/Dockerfile .

.PHONY: push
push:
	@docker push psycore/grpc-practice-server

.PHONY: upgrade
upgrade:
	@gsed -r -i "s/tag: [0-9]+/tag: $(TAG)/g" ./server/grpc-practice-server/values.yaml
	@gsed -r -i "s/version: [0-9]+/version: $(TAG)/g" ./server/grpc-practice-server/Chart.yaml
	@helm upgrade --install --force grpc-practice-server ./server/grpc-practice-server/

