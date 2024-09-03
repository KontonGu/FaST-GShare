DOCKER_USER=docker.io/kontonpuku666

.PHONY: clean_crd_gen
clean_crd_gen:
	rm -r pkg/client && rm -r pkg/apis/fastgshare.caps.in.tum/v1/zz_generated.deepcopy.go

.PHONY: update_crd
update_crd:
	bash code-gen.sh

.PHONY: code_gen_crd
code_frist_gen_crd:
	cd .. && git clone https://github.com/kubernetes/code-generator.git && cd code-generator && git checkout release-1.23
	bash code-gen.sh


### create dummPod_container
.PHONY: dummyPod_container
dummyPod_container:
	docker build -t ${DOCKER_USER}/dummycontainer:release -f docker/dummyPod/Dockerfile .


### ------------------------ fastpod-controller-manager ----------------------- ###
### create fastpod-controller-manager and corresponding container image
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

.PHONY: build-fastpod-controller-manager-container
build-fastpod-controller-manager-container:
	docker build -t ${DOCKER_USER}/fastpod-controller-manager:release -f docker/fastpod-controller-manager/Dockerfile .

.PHONY: test-build-fastpod-controller-manager-container
test-build-fastpod-controller-manager-container:
	docker build -t ${DOCKER_USER}/fastpod-controller-manager:controller_test -f docker/fastpod-controller-manager/Dockerfile .
	docker push ${DOCKER_USER}/fastpod-controller-manager:controller_test 
	# sudo ctr -n k8s.io i rm ${DOCKER_USER}/fastpod-controller-manager:controller_test
	sudo ctr -n k8s.io i ls | grep ${DOCKER_USER}/fastpod-controller-manager | awk '{print $$1}' | xargs -I {} sudo ctr -n k8s.io i rm {}

.PHONY: upload-fastpod-controller-manager-image
upload-fastpod-controller-manager-image:
	docker push ${DOCKER_USER}/fastpod-controller-manager:release 

.PHONY: clean-ctr-fastpod-controller-manager-image
clean-ctr-fastpod-controller-manager-image:
	sudo ctr -n k8s.io i rm ${DOCKER_USER}/fastpod-controller-manager:release


### ------------------------  fast-configurator ------------------------- ###
### create fastpod-controller-manager and corresponding container image
FAST_CONFIG_OUTPUT_DIR := cmd/fast-configurator
FAST_CONFIG_BUILD_DIR := cmd/fast-configurator
FAST_CONFIG_BINARY_NAME := fast-configurator

.PHONY: fast-configurator
fast-configurator:
	@echo "Building module [fast-configurator] ..."
	@cd $(FAST_CONFIG_BUILD_DIR) && go build -o $(FAST_CONFIG_BINARY_NAME)
	@echo "Build complete. Binary is located at $(FAST_CONFIG_OUTPUT_DIR)/$(FAST_CONFIGBINARY_NAME)"

.PHONY: test-fast-configurator-container
test-fast-configurator-container: build-fast-configurator-container upload-fast-configurator-image clean-ctr-fast-configurator-image

.PHONY: build-fast-configurator-container
build-fast-configurator-container:
	docker build -t ${DOCKER_USER}/fast-configurator:release -f docker/fast-configurator/Dockerfile .

.PHONY: upload-fast-configurator-image
upload-fast-configurator-image:
	docker push ${DOCKER_USER}/fast-configurator:release 

.PHONY: clean-ctr-fast-configurator-image
clean-ctr-fast-configurator-image:
	sudo ctr -n k8s.io i rm ${DOCKER_USER}/fast-configurator:release


# ----------------------------------------- debug fast-configurator and fastpodcontrollermanager ---------------------
.PHONY: test_fastconfigurator_fastpodcontrollermanager
test_fastconfigurator_fastpodcontrollermanager: clean-ctr-fast-configurator-image build-fast-configurator-container upload-fast-configurator-image \
clean-ctr-fastpod-controller-manager-image build-fastpod-controller-manager-container upload-fastpod-controller-manager-image



## --------------------------------------- openfaas fast-gshare dockerfile build ---------------------------------------------------
FAST_GSHARE_IMAGE_NAME := fast-gshare-faas
FAST_GSHARE_IMAGE_TAG=test
.PHONY: build-fast-gshare-faas-image
build-fast-gshare-faas-image:
	@echo "Building module [fast-gshare-faas] ..."
	docker build -t ${DOCKER_USER}/${FAST_GSHARE_IMAGE_NAME}:${FAST_GSHARE_IMAGE_TAG} -f Dockerfile .

.PHONY: upload-fast-gshare-faas-image
upload-fast-gshare-faas-image:
	docker push ${DOCKER_USER}/${FAST_GSHARE_IMAGE_NAME}:${FAST_GSHARE_IMAGE_TAG}

.PHONY: clean-fast-gshare-faas-image
clean-fast-gshare-faas-image:
	sudo ctr -n k8s.io i ls | grep -i ${FAST_GSHARE_IMAGE_NAME} | awk '{print $$1}' | xargs -I {} sudo ctr -n k8s.io i rm {} 

.PHONY: test-fast-gshare-faas-image
test-fast-gshare-faas-image: clean-fast-gshare-faas-image build-fast-gshare-faas-image upload-fast-gshare-faas-image
	




##------------------------------------ helm install the fast-gshare-fn system -------------------------------- ##
.PHONY: helm_install_fast-gshare-fn
helm_install_fast-gshare-fn:
	helm install fast-gshare ./chart/fastgshare --namespace fast-gshare --set functionNamespace=fast-gshare-fn  \
	--set  fastpodControllerManager.image="docker.io/kontonpuku666/fastpod-controller-manager:controller_test"

.PHONY: helm_uninstall_fast-gshare-fn
helm_uninstall_fast-gshare-fn:
	helm uninstall fast-gshare --namespace fast-gshare
	kubectl delete pod -l fastgshare/role=dummyPod -n kube-system
