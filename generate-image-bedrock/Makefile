.PHONY: build-BedrockImageGenerationFunction package-BedrockImageGenerationFunction

build-BedrockImageGenerationFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-BedrockImageGenerationFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap