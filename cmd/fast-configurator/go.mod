module fast-configurator

go 1.22.0

toolchain go1.22.5

replace github.com/KontonGu/FaST-GShare => ../../

require (
	github.com/KontonGu/FaST-GShare v0.0.0-00010101000000-000000000000
	github.com/NVIDIA/go-nvml v0.12.4-0
	k8s.io/klog/v2 v2.130.1
)

require github.com/go-logr/logr v1.4.1 // indirect
