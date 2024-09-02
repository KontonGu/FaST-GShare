/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.
@Techinical University of Munich, CAPS Cloud Team

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
	"flag"
	"log"
	"time"

	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	informersv1 "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions/fastgshare.caps.in.tum/v1"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/server"
	gcache "github.com/patrickmn/go-cache"

	"github.com/KontonGu/FaST-GShare/pkg/config"
	"github.com/KontonGu/FaST-GShare/pkg/controller"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/k8s"
	providertypes "github.com/openfaas/faas-provider/types"
	kubeinformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	"k8s.io/sample-controller/pkg/signals"

	informers "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
)

func main() {
	var kubeconfig string
	var masterURL string
	var verbose bool

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&verbose, "verbose", false, "Print")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kuebconfig. Only requeired if out-of-cluster.")

	flag.Parse()

	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())

	}

	kubeconfigQPS := 100
	kubeconfigBurst := 250

	clientCmdConfig.QPS = float32(kubeconfigQPS)
	clientCmdConfig.Burst = kubeconfigBurst

	kubeclient, err := kubernetes.NewForConfig(clientCmdConfig)
	if err != nil {
		log.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	fastpodClient, err := clientset.NewForConfig(clientCmdConfig)
	if err != nil {
		log.Fatalf("Error building fastpod clientset: %s", err.Error())
	}

	readConfig := config.ReadConfig{}

	osEnv := providertypes.OsEnv{}
	config, err := readConfig.Read(osEnv)

	if err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	config.Fprint(verbose)

	deployConfig := k8s.DeploymentConfig{
		RuntimeHTTPPort: 8080,
		HTTPProbe:       config.HTTPProbe,

		//SetNonRootUser:  config.SetNonRootUser,
		ReadinessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(config.ReadinessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(config.ReadinessProbeTimeoutSeconds),
			PeriodSeconds:       int32(config.ReadinessProbePeriodSeconds),
		},
		LivenessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(config.LivenessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(config.LivenessProbeTimeoutSeconds),
			PeriodSeconds:       int32(config.LivenessProbePeriodSeconds),
		},
		ImagePullPolicy:   config.ImagePullPolicy,
		ProfilesNamespace: config.ProfilesNamespace,
	}

	//the sync interval does not affect the scale to/from

	defaultResync := time.Minute * 5
	namespaceScope := config.DefaultFunctionNamespace

	if config.ClusterRole {
		namespaceScope = ""
	}

	kubeInformerOpt := kubeinformers.WithNamespace(namespaceScope)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeclient, defaultResync, kubeInformerOpt)

	fastpodInformerOpt := informers.WithNamespace(namespaceScope)
	fastpodInformerFactory := informers.NewSharedInformerFactoryWithOptions(fastpodClient, defaultResync, fastpodInformerOpt)

	factory := k8s.NewFunctionFactory(kubeclient, deployConfig)

	setup := serverSetup{
		config:                 config,
		functionFactory:        factory,
		kubeInformerFactory:    kubeInformerFactory,
		fastpodInformerFactory: fastpodInformerFactory,
		kubeClient:             kubeclient,
		fastpodClient:          fastpodClient,
	}

	klog.Info("Starting operator")
	runOperator(setup, config)
}

func runOperator(setup serverSetup, cfg config.BootstrapConfig) {
	kubeClient := setup.kubeClient
	fastpodClient := setup.fastpodClient
	kubeInformerFactory := setup.kubeInformerFactory
	fastpodInformerFactory := setup.fastpodInformerFactory

	factory := controller.FunctionFactory{
		Factory: setup.functionFactory,
	}

	setupLogging()

	stopCh := signals.SetupSignalHandler()

	//operator := true
	listers := startInformers(setup, stopCh)

	data := gcache.New(6*time.Minute, 10*time.Minute)
	fastpodLookup := k8s.NewFunctionLookup("fast-gshare-fn", listers.PodsInformer.Lister(), listers.FastPodsInformer.Lister(), data)

	rateScale := false

	if cfg.FunctionRate {
		rateScale = cfg.FunctionRate
	}

	klog.Infof("rate scale: %t", rateScale)

	fastpodLookup.RateRep = rateScale

	ctr := controller.NewController(
		kubeClient,
		fastpodClient,
		kubeInformerFactory,
		fastpodInformerFactory,
		factory,
		fastpodLookup,
	)

	srv := server.New(fastpodClient, kubeClient, listers.PodsInformer, listers.FastPodsInformer.Lister(), cfg.ClusterRole, fastpodLookup, cfg)

	go srv.Start()

	if err := ctr.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

type customInformers struct {
	//EndpointsInformer v1core.EndpointsInformer
	//DeploymentInformer v1apps.DeploymentInformer
	FastPodsInformer informersv1.FaSTPodInformer
	PodsInformer     kubeinformersv1.PodInformer
}

func startInformers(setup serverSetup, stopCh <-chan struct{}) customInformers {
	kubeInformerFactory := setup.kubeInformerFactory
	fastpodInformerFactory := setup.fastpodInformerFactory

	fastpods := fastpodInformerFactory.Fastgshare().V1().FaSTPods()
	go fastpods.Informer().Run(stopCh)

	if ok := cache.WaitForNamedCacheSync("fast-gshare:fastpods", stopCh, fastpods.Informer().HasSynced); !ok {
		klog.Fatal("Error failed to wait for cache to sync fastpods")
	}

	pods := kubeInformerFactory.Core().V1().Pods()
	go pods.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync("fast-gshare:pods", stopCh, pods.Informer().HasSynced); !ok {
		klog.Fatal("Error failed to wait for cache to sync pods")
	}

	return customInformers{
		PodsInformer:     pods,
		FastPodsInformer: fastpods,
	}

}

type serverSetup struct {
	config                 config.BootstrapConfig
	kubeClient             *kubernetes.Clientset
	fastpodClient          *clientset.Clientset
	functionFactory        k8s.FunctionFactory
	kubeInformerFactory    kubeinformers.SharedInformerFactory
	fastpodInformerFactory informers.SharedInformerFactory
	//profileInformerFactory informers.SharedInformerFactory
}

// paerse the command and set the corresponding flag of the klog
func setupLogging() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	//Sync
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)

		if f2 != nil {
			value := f1.Value.String()
			_ = f2.Value.Set(value)
		}
	})
}
