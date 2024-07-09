exec_name="get_uuid"

CUDA_PATH ?= /usr/local/cuda-12.2
CUDA_INCLUDE += -I$(CUDA_PATH)/include -I./

build: get_uuid.cpp
	g++ -o ${exec_name} ${CUDA_INCLUDE} get_uuid.cpp -lnvidia-ml 
clean:
	rm ${exec_name}