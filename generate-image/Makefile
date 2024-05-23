.PHONY: build-ImageGenerationFunction package-ImageGenerationFunction

build-ImageGenerationFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-ImageGenerationFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap