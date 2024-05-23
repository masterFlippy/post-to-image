.PHONY: build-PromptPreparationAPIFunction package-PromptPreparationAPIFunction

build-PromptPreparationAPIFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-PromptPreparationAPIFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap