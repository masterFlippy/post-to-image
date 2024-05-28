.PHONY: build-DalleImageGenerationFunction package-DalleImageGenerationFunction

build-DalleImageGenerationFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-DalleImageGenerationFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap