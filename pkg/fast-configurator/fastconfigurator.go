/*
 * Created/Modified on Mon May 27 2024
 *
 * Author: KontonGu
 * Copyright (c) 2024 TUM - CAPS Cloud
 * Licensed under the Apache License, Version 2.0 (the "License")
 */

package fastconfigurator

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"k8s.io/klog/v2"
)

var (
	GPUClientsIPEnv = "FaSTPod-GPUClientsIP"
)

const (
	GPUClientsIPFile       = "/fastpod/library/GPUClientsIP.txt"
	FastSchedulerConfigDir = "/fastpod/scheduler/config"
	GPUClientsPortConfigDir    = "/fastpod/scheduler/gpu_clients"
)

func Run(deviceCtrManager string) {
	gcIpFile, err := os.Create(GPUClientsIPFile);
	err != nil {
		klog.Errorf("Error Cannot create the GPUClientsIPFile = %s\n", GPUClientsIPFile)
	}
	st := os.Getenv(GPUClientsIPEnv) + "\n"
	gcIpFile.WriteString(st)
	gcIpFile.Sync()
	gcIpFile.Close()

	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Error Cannot get hostname. \n")
		panic(err)
	}

	os.MkdirAll(FastSchedulerConfigDir, os.ModePerm)
	os.MkdirAll(GPUClientsConfigDir, os.ModePerm)

	klog.Info("Trying to connet controller-manager....., server IP:Port = %s\n", deviceCtrManager)
	conn, err := net.Dial("tcp", deviceCtrManager)
	if err != nil{
		klog.Fatalf("Error Cannot connect to the device-controller-manager, IP:Port = %s .", deviceCtrManager)
		panic(err)
	}

	klog.Info("Connection to the device-controller-manager succeed.")
	reader := bufio.NewReader(conn)

	writeMsgToConn(conn, "hostname:"+hostname+"\n")
	registerGPUDevices(conn)
	
	nodeHeartbeater := time.NewTicker(time.Second * 15)
	go sendNodeHeartbeats(conn, nodeHeartbeater.C)

	recvMsgDeviceCtrManager(reader)
}


func writeMsgToConn(conn net.Conn, st string) error {
	_, err := conn.Write([]byte(st))
	if err != nil {
		klog.Errorf("Error failed to write msg: %s\n", st)
		return err
	}
	return nil
}

func registerGPUDevices(conn net.Conn){
	gpu_num, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		klog.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}
	var buf bytes.Buffer
	for i := 0; i < gpu_num; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		meminfo, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get GPU memory %d: %v", i, nvml.ErrorString(ret))
		}
		memsize := meminfo.Total

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			klog.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		buf.WriteString(uuid+":")
		buf.WriteString(strconv.FormatUint(memsize, 10))
		buf.WriteString(",")
	}
	buf.WriteString("\n")
	klog.Infof("Sucessfully get GPU devices, <uuid>:<memory>,... = %s\n", buf.String())
	conn.Write(buf.Bytes())

}

func sendNodeHeartbeats(conn net.Conn, heartTick <-chan){
	klog.Infof("Send node heartbeat to fastpod-controller-manager: %s\n", time.Now().String())
	writeMsgToConn(conn, "heartbeat!\n")
	for {
		<- heartTick
		klog.Infof("Send node heartbeat to fastpod-controller-manager: %s\n", time.Now().String())
		writeMsgToConn(conn, "heartbeat!\n")
	}
}

func recvMsgFastCtrManager(reader *bufio.Reader){
	klog.Infof("Receiving Resource and Port Configuration from fastpod-controller-manager. \n")
	for {
		configMsg, err := reader.ReadString('\n')
		if err != nil {
			klog.Errorf("Error while Receiving Msg from fastpod-controller-manager")
			return
		}
		handleMsg(configMsg)
	}
}

func handleMsg(msg string){
	klog.Infof("Received message from fastpod-controller-manager: %s", msg)
	validMsg := string(msg[:len(msg)-1])
	// write configuration to the scheudler configuration path
	// original message uses ":" to separate field, "," to separate items of the field 
	msgParsed := strings.Split(validMsg, ":")
	if len(msgParsed) != 3 {
		klog.Errorf("Error Received wrong format of configuration msg: %s\n", validMsg)
		return
	}
	uuid, fastSchedConf, gpuClientsPort := msgParsed[0], msgParsed[1], msgParsed[2]
	
	confPath := filepath.Join(FastSchedulerConfigDir, uuid)
	confFile, err := os.Create(confPath)
	if err != nil {
		klog.Errorf("Error failed to create the fast scheduler resource configuration file: %s\n.", confPath)
	}

	gcPortFilePath := filepath.Join(GPUClientsPortConfigDir, uuid)
	gcPortFile, err := os.Create(gcPortFilePath)
	if err != nil {
		klog.Errorf("Error failed to create the gpu clients' port configuration file: %s\n.", confPath)
	}
	
	confFile.WriteString(fmt.Sprintf("%d\n", strings.Count(fastSchedConf, ",")))
	confFile.WriteString(strings.ReplaceAll(fastSchedConf, ",", "\n"))
	confFile.Sync()
	confFile.Close()

	gcPortFile.WriteString(fmt.Sprintf("%d\n", strings.Count(gpuClientsPort, ",")))
	gcPortFile.WriteString(strings.ReplaceAll(gpuClientsPort, ",", "\n"))
	gcPortFile.Sync()
	gcPortFile.Close()
}