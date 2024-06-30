/*
 * Created/Modified on Mon May 27 2024
 *
 * Author: KontonGu
 * Copyright (c) 2024 TUM - CAPS Cloud
 * Licensed under the Apache License, Version 2.0 (the "License")
 */

package fastconfigurator

import "fmt"

const (
	ClientsIPFile          = "/fastpod/library/GPUClientIP.txt"
	FastSchedulerConfigDir = "/fastpod/scheduler/config"
	GPUClientsConfigDir    = "/fastpod/scheduler/gpu_clients"
)

func Run(device_manager string) {
	fmt.Println("Run function testing")
}
