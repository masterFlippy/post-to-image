.PHONY: build-GetImageFunction package-GetImageFunction

build-GetImageFunction:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
	chmod +x bootstrap
	cp bootstrap $(ARTIFACTS_DIR)/

package-GetImageFunction:
	cd $(ARTIFACTS_DIR) && zip -r function.zip bootstrap