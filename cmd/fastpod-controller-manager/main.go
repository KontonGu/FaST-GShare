/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.

Implemntation is based on k8s sample-controller:
https://github.com/kubernetes/sample-controller/tree/master

Copyright (c) 2024 TUM - CAPS Cloud
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	informers "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions"
	fastpodcontrollermanager "github.com/KontonGu/FaST-GShare/pkg/fastpod-controller-manager"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	"k8s.io/sample-controller/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
	workerNum  int
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the shutdown signal gracefully
	ctx := signals.SetupSignalHandler()
	// logger := klog.FromContext(ctx)

	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube/config")
	}

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Error("Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	cfg.QPS = 512.0
	cfg.Burst = 512

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Error("Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	fastpodClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	crdExist := checkCRDExist(fastpodClient)
	if !crdExist {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	fastpodInformerFactory := informers.NewSharedInformerFactory(fastpodClient, time.Second*30)

	controller := fastpodcontrollermanager.NewController(ctx, kubeClient, fastpodClient,
		kubeInformerFactory.Core().V1().Nodes(),
		kubeInformerFactory.Core().V1().Pods(),
		fastpodInformerFactory.Fastgshare().V1().FaSTPods())

	kubeInformerFactory.Start(ctx.Done())
	fastpodInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, workerNum); err != nil {
		klog.Error("Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&workerNum, "workers", 2, "The number of workers for handler in reconcile.")
}

func checkCRDExist(fastpodClientset *clientset.Clientset) bool {
	_, err := fastpodClientset.FastgshareV1().FaSTPods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		if _, ok := err.(*errors.StatusError); ok {
			if errors.IsNotFound(err) {
				klog.Fatalf("The FaSTPod CRD is still not created.")
				return false
			}
		}
	}
	return true
}
