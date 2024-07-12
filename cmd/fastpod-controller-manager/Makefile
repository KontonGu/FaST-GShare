CTR_MGR_OUTPUT_DIR := ./
CTR_MGR_BUILD_DIR := ./
CTR_MGR_BINARY_NAME := fastpodcontrollermanager

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(CTR_MGR_OUTPUT_DIR)/$(CTR_MGR_BINARY_NAME)
	@echo "Clean complete."

.PHONY: build
build:
	@echo "Building project..."
	@cd $(CTR_MGR_BUILD_DIR) && go build -o $(CTR_MGR_BINARY_NAME)
	@echo "Build complete. Binary is located at $(CTR_MGR_OUTPUT_DIR)/$(CTR_MGR_BINARY_NAME)"