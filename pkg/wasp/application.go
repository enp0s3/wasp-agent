/*
 * This file is part of the Wasp project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023,Red Hat, Inc.
 *
 */
package wasp

import (
	"context"
	"flag"
	"fmt"
	"github.com/openshift-virtualization/wasp-agent/pkg/client"
	"github.com/openshift-virtualization/wasp-agent/pkg/informers"
	"github.com/openshift-virtualization/wasp-agent/pkg/log"
	eviction_controller "github.com/openshift-virtualization/wasp-agent/pkg/wasp/eviction-controller"
	limited_swap_manager "github.com/openshift-virtualization/wasp-agent/pkg/wasp/limited-swap-manager"
	shortage_detector "github.com/openshift-virtualization/wasp-agent/pkg/wasp/shortage-detector"
	stats_collector "github.com/openshift-virtualization/wasp-agent/pkg/wasp/stats-collector"
	"github.com/shirou/gopsutil/mem"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"time"
)

type WaspApp struct {
	podStatsCollector               stats_collector.PodStatsCollector
	evictionController              *eviction_controller.EvictionController
	limitesSwapManager              *limited_swap_manager.LimitedSwapManager
	podInformer                     cache.SharedIndexInformer
	nodeInformer                    cache.SharedIndexInformer
	ctx                             context.Context
	maxMemoryOverCommitmentBytes    resource.Quantity
	cli                             client.WaspClient
	maxAverageSwapInPagesPerSecond  float32
	maxAverageSwapOutPagesPerSecond float32
	swapUtilizationThresholdFactor  float64
	AverageWindowSizeSeconds        time.Duration
	waspNs                          string
	nodeName                        string
	fsRoot                          string
	totalSwapMemoryBytes            uint64
	totalMemoryBytes                uint64
}

func Execute() {
	var err error
	flag.Parse()

	setCrioSocketSymLink()
	setOCIHook()

	var app = WaspApp{}
	memoryOverCommitmentThreshold := os.Getenv("MEMORY_OVER_COMMITMENT_THRESHOLD")
	maxAverageSwapInPagesPerSecond := os.Getenv("MAX_AVERAGE_SWAP_IN_PAGES_PER_SECOND")
	maxAverageSwapOutPagesPerSecond := os.Getenv("MAX_AVERAGE_SWAP_OUT_PAGES_PER_SECOND")
	AverageWindowSizeSeconds := os.Getenv("AVERAGE_WINDOW_SIZE_SECONDS")
	swapUtilizationThresholdFactor := os.Getenv("SWAP_UTILIZATION_THRESHOLD_FACTOR")
	app.nodeName = os.Getenv("NODE_NAME")
	app.fsRoot = os.Getenv("FSROOT")

	app.podStatsCollector = stats_collector.NewPodSummaryCollector()
	err = app.podStatsCollector.Init()
	if err != nil {
		panic(err)
	}

	app.maxMemoryOverCommitmentBytes, err = resource.ParseQuantity(memoryOverCommitmentThreshold)
	if err != nil {
		panic(err)
	}

	AverageWindowSizeSecondsToConvert, err := strconv.Atoi(AverageWindowSizeSeconds)
	if err != nil {
		panic(err)
	}
	app.AverageWindowSizeSeconds = time.Duration(AverageWindowSizeSecondsToConvert) * time.Second

	maxAverageSwapInPagesPerSecondToConvert, err := strconv.Atoi(maxAverageSwapInPagesPerSecond)
	if err != nil {
		panic(err)
	}
	app.maxAverageSwapInPagesPerSecond = float32(maxAverageSwapInPagesPerSecondToConvert)

	maxSwapOutRateToConvert, err := strconv.Atoi(maxAverageSwapOutPagesPerSecond)
	if err != nil {
		panic(err)
	}
	app.maxAverageSwapOutPagesPerSecond = float32(maxSwapOutRateToConvert)

	utilizationFactor, err := strconv.ParseFloat(swapUtilizationThresholdFactor, 64)
	if err != nil {
		panic(err)
	}
	app.swapUtilizationThresholdFactor = float64(utilizationFactor)

	memStat, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	app.totalMemoryBytes = memStat.Total

	swapMemStat, err := mem.SwapMemory()
	if err != nil {
		panic(err)
	}
	app.totalSwapMemoryBytes = swapMemStat.Total

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.ctx = ctx

	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	app.waspNs = string(nsBytes)
	app.cli, err = client.GetWaspClient()
	if err != nil {
		panic(err)
	}
	app.podInformer = informers.GetPodInformer(app.cli)
	app.nodeInformer = informers.GetNodeInformer(app.cli)

	log.Log.Infof("MEMORY_OVER_COMMITMENT_THRESHOLD:%v "+
		"MAX_AVERAGE_SWAP_IN_PAGES_PER_SECOND:%v "+
		"MAX_AVERAGE_SWAP_OUT_PAGES_PER_SECOND:%v "+
		"AVERAGE_WINDOW_SIZE_SECONDS:%v "+
		"SWAP_UTILIZATION_THRESHOLD_FACTOR: %v"+
		"nodeName: %v "+
		"ns: %v "+
		"fsRoot: %v",
		app.maxMemoryOverCommitmentBytes,
		app.maxAverageSwapInPagesPerSecond,
		app.maxAverageSwapOutPagesPerSecond,
		app.AverageWindowSizeSeconds,
		app.swapUtilizationThresholdFactor,
		app.nodeName,
		app.waspNs,
		app.fsRoot,
	)

	stop := ctx.Done()
	app.initEvictionController(stop)
	app.initLimitedSwapManager(stop)
	app.Run(stop)
}

func (waspapp *WaspApp) initEvictionController(stop <-chan struct{}) {
	sc := stats_collector.NewStatsCollectorImpl()
	shortageDetector := shortage_detector.NewShortageDetectorImpl(sc,
		waspapp.maxAverageSwapInPagesPerSecond,
		waspapp.maxAverageSwapOutPagesPerSecond,
		waspapp.swapUtilizationThresholdFactor,
		waspapp.maxMemoryOverCommitmentBytes.Value(),
		waspapp.AverageWindowSizeSeconds,
		waspapp.totalMemoryBytes,
		waspapp.totalSwapMemoryBytes)

	waspapp.evictionController = eviction_controller.NewEvictionController(waspapp.cli,
		waspapp.podStatsCollector,
		sc,
		shortageDetector,
		waspapp.podInformer,
		waspapp.nodeInformer,
		waspapp.nodeName,
		waspapp.waspNs,
		stop)
}

func (waspapp *WaspApp) initLimitedSwapManager(stop <-chan struct{}) {
	waspapp.limitesSwapManager = limited_swap_manager.NewLimitedSwapManager(waspapp.cli,
		waspapp.podInformer,
		waspapp.nodeName,
		stop,
	)
}

func (waspapp *WaspApp) Run(stop <-chan struct{}) {
	go waspapp.podInformer.Run(stop)
	go waspapp.nodeInformer.Run(stop)

	if !cache.WaitForCacheSync(stop,
		waspapp.podInformer.HasSynced,
		waspapp.nodeInformer.HasSynced,
	) {
		klog.Warningf("failed to wait for caches to sync")
	}

	go func() {
		waspapp.evictionController.Run(waspapp.ctx)
	}()
	go func() {
		waspapp.limitesSwapManager.Run(1)
	}()

	<-waspapp.ctx.Done()

}

func setCrioSocketSymLink() {
	err := os.MkdirAll("/var/run/crio", 0755)
	if err != nil {
		klog.Warningf(err.Error())
		return
	}
	os.Symlink("/host/var/run/crio/crio.sock", "/var/run/crio/crio.sock")
	if err != nil {
		klog.Warningf(err.Error())
		return
	}
}

func setOCIHook() {
	err := moveFile("/app/OCI-hook/hook.sh", "/host/opt/oci-hook-swap.sh")
	if err != nil {
		klog.Warningf(err.Error())
		return
	}
	err = moveFile("/app/OCI-hook/swap-for-burstable.json", "/host/run/containers/oci/hooks.d/swap-for-burstable.json")
	if err != nil {
		klog.Warningf(err.Error())
	}
}

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Couldn't open dest file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("Couldn't copy to dest from source: %v", err)
	}

	inputFile.Close()

	// Set file permissions to make it executable
	err = os.Chmod(destPath, 0755)
	if err != nil {
		return fmt.Errorf("Couldn't set file permissions: %v", err)
	}

	return nil
}
