.PHONY: build-ImageUploadFunction package-ImageUploadFunction

build-ImageUploadFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-ImageUploadFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap