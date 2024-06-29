package main

import (
	"flag"

	// "github.com/KontonGu/FaST-GShare/pkg/fast-configurator/fastconfigurator"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	klog "k8s.io/klog/v2"
)

var (
	device_manager_ip_port string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		klog.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
		select {}
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	gpu_num, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		klog.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}
	klog.Infof("The number of gpu devices: %d\n", gpu_num)
	for i := 0; i < gpu_num; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		klog.Infof("%v\n", uuid)
	}
	// fastconfigurator.Run(device_manager_ip_port)

}

func init() {
	flag.StringVar(&device_manager_ip_port, "devicemanager_ip_port", "127.0.0.1:9797", "The IP and Port to the FaST-GShare device manager.")
}
