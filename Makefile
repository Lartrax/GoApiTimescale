GPC_PROJECT_ID=min-backend
SERVICE_NAME=go-api-postgres
CONTAINER_NAME=eu.gcr.io/$(GPC_PROJECT_ID)/$(SERVICE_NAME)

run: build
	docker run -p 8080:8080 $(CONTAINER_NAME)
build:
	docker build -t $(CONTAINER_NAME) .
push: build
	docker push $(CONTAINER_NAME)
deploy: build push
	gcloud run deploy $(SERVICE_NAME)\
		--project $(GPC_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
test:
	go test ./...