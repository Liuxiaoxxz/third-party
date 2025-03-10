package jvmxexporter

import (
	"context"
	"github.com/Liuxiaoxxz/third-party/grpc/metrics"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"strings"
)

var (
	dict = map[string]string{
		// MemoryPool 映射
		"G1 Survivor Space":                "G1SurvivorSpace",
		"G1 Eden Space":                    "G1EdenSpace",
		"Compressed Class Space":           "CompressedClassSpace",
		"CodeHeap 'non-nmethods'":          "CodeHeap'non-nmethods'",
		"Metaspace":                        "Metaspace",
		"CodeHeap 'non-profiled nmethods'": "CodeHeap'non-profilednmethods'",
		"G1 Old Gen":                       "G1OldGen",
		// GC 映射
		"G1 Young Generation": "G1 Young Generation",
		"G1 Old Generation":   "G1 Old Generation",
		"G1 Concurrent GC":    "G1 Concurrent GC",
	}
)

const (
	JVM_MEMORY_USED      = "jvm.memory.used"
	JVM_MEMORY_COMMITTED = "jvm.memory.committed"
	JVM_MEMORY_LIMITI    = "jvm.memory.limit"
	JVM_GC_DURATION      = "jvm.gc.duration"
	JVM_THREAD_COUNT     = "jvm.thread.count"

	JVM_GC_NAME          = "jvm.gc.name"
	JVM_THREAD_DAEMON    = "jvm.thread.daemon"
	JVM_MEMORY_POOL_NAME = "jvm.memory.pool.name"
)

func metricTransform(ctx context.Context, md pmetric.Metrics) (*metrics.ExportMetricsServiceRequest, error) {
	data := &metrics.ExportMetricsServiceRequest{}
	resourceMetrics := md.ResourceMetrics()
	rmsLen := resourceMetrics.Len()
	for i := 0; i < rmsLen; i++ {
		resourceMetric := resourceMetrics.At(i)
		resource := resourceMetric.Resource()
		resourceAttributes := resource.Attributes()
		if appname, b := resourceAttributes.Get("service.name"); b == true {
			data.AppName = appname.AsString()
		}
		if pid, b := resourceAttributes.Get("process.pid"); b == true {
			data.Pid = string(pid.Int())
		}
		scopeMetrics := resourceMetric.ScopeMetrics()
		smsLen := scopeMetrics.Len()
		for i := 0; i < smsLen; i++ {
			scopeMetric := scopeMetrics.At(i)
			scopeName := scopeMetric.Scope().Name()
			if strings.Contains(scopeName, "io.opentelemetry.runtime-telemetry-java") {
				metrics := scopeMetric.Metrics()
				msLen := metrics.Len()
				for i := 0; i < msLen; i++ {
					metric := metrics.At(i)
					copeMetric(data, metric)
				}
			}
		}
	}

	return data, nil
}

func copeMetric(data *metrics.ExportMetricsServiceRequest, metric pmetric.Metric) {
	// 确保 MemoryPool 和 GarbageCollector 的 maps 被初始化
	if data.MemoryPool.MemoryUsages == nil {
		data.MemoryPool.MemoryUsages = make(map[string]*metrics.MemoryUsage)
	}
	if data.GarbageCollector.GarbageCollectors == nil {
		data.GarbageCollector.GarbageCollectors = make(map[string]*metrics.GarbageCollectorInfo)
	}

	switch metric.Name() {
	case JVM_MEMORY_USED, JVM_MEMORY_COMMITTED, JVM_MEMORY_LIMITI:
		dataPoints := metric.Sum().DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			dataPoint := dataPoints.At(i)
			dataPointAttributes := dataPoint.Attributes()
			v, b := dataPointAttributes.Get(JVM_MEMORY_POOL_NAME)
			if b {
				poolName := v.AsString()
				// 获取池名映射
				mappedName := dict[poolName]
				if mappedName == "" {
					// 如果映射不到，使用 poolName 本身作为默认值
					mappedName = poolName
				}
				// 确保内存池数据被初始化
				if _, exists := data.MemoryPool.MemoryUsages[mappedName]; !exists {
					data.MemoryPool.MemoryUsages[mappedName] = &metrics.MemoryUsage{
						Max: -1,
					}
				}
				// 根据字段填充数据
				memoryUsage := data.MemoryPool.MemoryUsages[mappedName]
				if metric.Name() == JVM_MEMORY_USED {
					memoryUsage.Used = dataPoint.IntValue()
				} else if metric.Name() == JVM_MEMORY_COMMITTED {
					memoryUsage.Committed = dataPoint.IntValue()
				} else if metric.Name() == JVM_MEMORY_LIMITI {
					memoryUsage.Max = dataPoint.IntValue()
				}
			}
		}
	case JVM_GC_DURATION:
		histogram := metric.Histogram()
		dataPoints := histogram.DataPoints()
		for i := 0; i < dataPoints.Len(); i++ {
			dataPoint := dataPoints.At(i)
			dataPointAttributes := dataPoint.Attributes()
			name, b := dataPointAttributes.Get(JVM_GC_NAME)
			if b {
				garbageCollectorName := dict[name.AsString()]
				if garbageCollectorName == "" {
					// 如果映射不到，使用 poolName 本身作为默认值
					garbageCollectorName = name.AsString()
				}
				// 确保垃圾收集器数据被初始化
				if _, exists := data.GarbageCollector.GarbageCollectors[garbageCollectorName]; !exists {
					data.GarbageCollector.GarbageCollectors[garbageCollectorName] = &metrics.GarbageCollectorInfo{}
				}
				garbageCollectorInfo := data.GarbageCollector.GarbageCollectors[garbageCollectorName]
				garbageCollectorInfo.Name = garbageCollectorName
				garbageCollectorInfo.CollectionCount = dataPoint.Count()
				garbageCollectorInfo.CollectionTime = int32(dataPoint.Sum() * 1000)
			}
		}
	case JVM_THREAD_COUNT:
		sum := metric.Sum()
		dataPoints := sum.DataPoints()
		var threadCount int64
		var daemonThreadCount int64
		for i := 0; i < dataPoints.Len(); i++ {
			dataPoint := dataPoints.At(i)
			currentThreadCount := dataPoint.IntValue()
			threadCount += currentThreadCount
			isDaemon, b := dataPoint.Attributes().Get(JVM_THREAD_DAEMON)
			if b && isDaemon.Bool() {
				daemonThreadCount += currentThreadCount
			}
		}
		data.Thread.ThreadCount = threadCount
		data.Thread.PeakThreadCount = threadCount
		data.Thread.DeamonThreadCount = daemonThreadCount
	}
}
