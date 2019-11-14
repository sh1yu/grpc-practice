.PHONY: server
server:
	docker build -t psycore/grpc-practice-server -f server/Dockerfile .