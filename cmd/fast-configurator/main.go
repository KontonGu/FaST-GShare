package main

import (
	"flag"
	"strings"

	fastconfigurator "github.com/KontonGu/FaST-GShare/pkg/fast-configurator"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	klog "k8s.io/klog/v2"
)

var (
	ctr_mgr_ip_port string
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

		meminfo, ret := device.GetMemoryInfo()
		memsize := meminfo.Total
		klog.Infof("Memory Size = %d\n", memsize)

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		klog.Infof("UUID: %v\n", uuid)

		gpu_type_name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get name of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		type_name := strings.Split(gpu_type_name, " ")[1]
		klog.Infof("GPU %d=%s\n", i, type_name)

	}
	fastconfigurator.Run(ctr_mgr_ip_port)

}

func init() {
	flag.StringVar(&ctr_mgr_ip_port, "ctr_mgr_ip_port", "fastpod-controller-manager-svc.kube-system.svc.cluster.local:10086", "The IP and Port to the FaST-GShare device manager.")
}
