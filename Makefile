.PHONY:
	build
	dev

build:
	docker build --no-cache -t rate-service .

dev:
	docker run -p 9000:9000 rate-service
