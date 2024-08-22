#include <iostream>
#include <nvml.h>
#include <unistd.h>
#include <sstream>

int main(){
    nvmlReturn_t result;
    unsigned int device_count;
    
    result = nvmlInit();
    if (result != NVML_SUCCESS) {
        std::cerr << "Failed to initialize NVML: " << nvmlErrorString(result) << std::endl;
        return 1;
    }

    // Get the number of GPU devices
    result = nvmlDeviceGetCount(&device_count);
    if (result != NVML_SUCCESS) {
        std::cerr << "Failed to get device count: " << nvmlErrorString(result) << std::endl;
        nvmlShutdown();
        return 1;
    }

     for (unsigned int i = 0; i < device_count; ++i) {
        nvmlDevice_t device;
        char deviceUUID[NVML_DEVICE_UUID_BUFFER_SIZE];
        // Get the handle for the device
        result = nvmlDeviceGetHandleByIndex(i, &device);
        if (result != NVML_SUCCESS) {
            std::cerr << "Failed to get handle for device " << i << ": " << nvmlErrorString(result) << std::endl;
            continue;
        }

        // Get the device UUID
        result = nvmlDeviceGetUUID(device, deviceUUID, NVML_DEVICE_UUID_BUFFER_SIZE);
        if (result != NVML_SUCCESS) {
            std::cerr << "Failed to get UUID for device " << i << ": " << nvmlErrorString(result) << std::endl;
            continue;
        }

         // Get the gpu name
        char gpu_name[NVML_DEVICE_NAME_BUFFER_SIZE];
        result = nvmlDeviceGetName(device, gpu_name, NVML_DEVICE_NAME_BUFFER_SIZE);
        if (result != NVML_SUCCESS) {
            std::cerr << "Failed to get name of device " << i << ": " << nvmlErrorString(result) << std::endl;
            continue;
        }
        std::stringstream ss(gpu_name); 
        std::string gpu_type,tmp;
        ss >> tmp;
        ss >> gpu_type;
        
        std::cout << deviceUUID << ":" << gpu_type << std::endl;
     }

    result = nvmlShutdown();
    if (result != NVML_SUCCESS) {
        std::cerr << "Failed to shutdown NVML: " << nvmlErrorString(result) << std::endl;
        return 1;
    }
    pause();
    return 0;

}