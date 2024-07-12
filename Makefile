.PHONY: clean_crd_gen
clean_crd_gen:
	rm -r pkg/client && rm -r pkg/apis/fastgshare.caps.in.tum/v1/zz_generated.deepcopy.go

.PHONY: update_crd
update_crd:
	bash code-gen.sh

.PHONY: code_gen_crd
code_gen_crd:
	cd .. && git clone https://github.com/kubernetes/code-generator.git && cd code-generator && git checkout release-1.23
	bash code-gen.sh

.PHONY: dummyPod_container
dummyPod_container:
	docker build -t docker.io/kontonpuku666/dummycontainer:release -f docker/dummyPod/Dockerfile .

.PHONY: fastpod-controller-manager-dummyPod-container
fastpod-controller-manager-container:
	docker build -t docker.io/kontonpuku666/fastpod-controller-manager:release -f docker/fastpod-controller-manager/Dockerfile .

CTR_MGR_OUTPUT_DIR := cmd/fastpod-controller-manager
CTR_MGR_BUILD_DIR := cmd/fastpod-controller-manager
CTR_MGR_BINARY_NAME := fastpodcontrollermanager

.PHONY: ctr_mgr_clean
ctr_mgr_clean:
	@echo "Cleaning up..."
	@rm -f $(CTR_MGR_OUTPUT_DIR)/$(CTR_MGR_BINARY_NAME)
	@echo "Clean complete."

.PHONY: fastpod-controller-manager
fastpod-controller-manager:
	@echo "Building module [fastpod-controller-manager] ..."
	@cd $(CTR_MGR_BUILD_DIR) && go build -o $(CTR_MGR_BINARY_NAME)
	@echo "Build complete. Binary is located at $(CTR_MGR_OUTPUT_DIR)/$(CTR_MGR_BINARY_NAME)"