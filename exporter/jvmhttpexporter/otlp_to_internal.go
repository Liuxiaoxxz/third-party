package jvmhttpexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"strings"
)

type Data struct {
	LogMessage LogMessage `json:"logMessage"`
	LogType    string     `json:"logType"`
	MasterIp   string     `json:"masterIp"`
}

type LogMessage struct {
	jManagementMessage JManagementMessage `json:"jManagementMessage"`
	apmLang            string             `json:"apm-lang"`
}

// 内部格式的数据结构
type JManagementMessage struct {
	BufferPool struct {
		Mapped struct {
			Count    int `json:"count"`
			Used     int `json:"used"`
			Capacity int `json:"capacity"`
		} `json:"mapped"`
		Direct struct {
			Count    int `json:"count"`
			Used     int `json:"used"`
			Capacity int `json:"capacity"`
		} `json:"direct"`
	} `json:"bufferPool"`
	AgentId                   string           `json:"agentId"`
	CreationTime              string           `json:"creationTime"`
	AppName                   string           `json:"appName"`
	AppStartTime              string           `json:"appStartTime"`
	CPU                       CPU              `json:"cpu"`
	Pid                       string           `json:"pid"`
	Thread                    Thread           `json:"thread"`
	MemoryPool                MemoryPool       `json:"memoryPool"`
	Version                   string           `json:"version"`
	Docker                    bool             `json:"docker"`
	GarbageCollector          GarbageCollector `json:"garbageCollector"`
	MultiAgentId              string           `json:"multiAgentId"`
	DatabaseConnectionMessage struct {
		LeakSuspicious                 []interface{} `json:"leakSuspicious"`
		DatabaseConnectionMessageArray []interface{} `json:"databaseConnectionMessageArray"`
	} `json:"databaseConnectionMessage"`
	Status int `json:"status"`
}

type CPU struct {
	ProcessCpu    float64 `json:"processCpu"`
	AvgSystemCpu  float64 `json:"avgSystemCpu"`
	SystemCpu     float64 `json:"systemCpu"`
	AvgProcessCpu float64 `json:"avgProcessCpu"`
}

type Thread struct {
	ThreadCount             int64       `json:"threadCount"`
	ThreadInfos             ThreadInfos `json:"threadInfos"`
	TotalStartedThreadCount int         `json:"totalStartedThreadCount"`
	PeakThreadCount         int64       `json:"peakThreadCount"`
	DeamonThreadCount       int64       `json:"deamonThreadCount"`
}

type ThreadInfos struct {
	ThreadInfo [][]interface{} `json:"threadInfo"`
	LockNames  []string        `json:"lockNames"`
}

type MemoryPool struct {
	MemoryUsages map[string]*MemoryUsage `json:"memoryUsages"`
}

type MemoryUsage struct {
	Init      int64 `json:"init"`
	Committed int64 `json:"committed"`
	Max       int64 `json:"max"`
	Used      int64 `json:"used"`
}

type GarbageCollector struct {
	GarbageCollectors map[string]*GarbageCollectorInfo `json:"garbageCollectors"`
}

type GarbageCollectorInfo struct {
	Valid           bool     `json:"valid"`
	CollectionTime  int      `json:"collectionTime"`
	MemoryPoolNames []string `json:"memoryPoolNames"`
	CollectionCount uint64   `json:"collectionCount"`
	Name            string   `json:"name"`
}

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

func metricTransform(ctx context.Context, md pmetric.Metrics) ([]byte, error) {
	data := &Data{}
	data.LogType = "JavaManagementData"
	data.MasterIp = "110.011.178.231,127.0.0.1"
	data.LogMessage = LogMessage{}
	logMessage := &data.LogMessage
	logMessage.jManagementMessage = JManagementMessage{}
	jManagementMessage := &logMessage.jManagementMessage
	resourceMetrics := md.ResourceMetrics()
	rmsLen := resourceMetrics.Len()
	for i := 0; i < rmsLen; i++ {
		resourceMetric := resourceMetrics.At(i)
		resource := resourceMetric.Resource()
		resourceAttributes := resource.Attributes()
		if appname, b := resourceAttributes.Get("service.name"); b == true {
			jManagementMessage.AppName = appname.AsString()
		}
		if pid, b := resourceAttributes.Get("process.pid"); b == true {
			jManagementMessage.Pid = string(pid.Int())
		}
		//mock的假数据
		jManagementMessage.AgentId = "bookdemo-bookdemo-746cc6d5f4-qmgsb-7d6e2b43-5c8e-49ba-8852-40a7fd8c0e8f-bookdemo-org.apache.catalina.startup.Bootstrap-1@192.168.136.105:8080"
		jManagementMessage.MultiAgentId = "bookdemo-bookdemo-746cc6d5f4-qmgsb-7d6e2b43-5c8e-49ba-8852-40a7fd8c0e8f-bookdemo-org.apache.catalina.startup.Bootstrap-1@192.168.136.105:8080%bookdemo-bookdemo-746cc6d5f4-qmgsb-7d6e2b43-5c8e-49ba-8852-40a7fd8c0e8f-bookdemo-org.apache.catalina.startup.Bootstrap-1@192.168.136.105:40195%bookdemo-bookdemo-746cc6d5f4-qmgsb-7d6e2b43-5c8e-49ba-8852-40a7fd8c0e8f-bookdemo-org.apache.catalina.startup.Bootstrap-1@192.168.136.105:36373"
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
					copeMetric(jManagementMessage, metric)
				}
			}
		}
	}
	// 将结构体转换为 JSON 字节数组
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err) // 处理序列化错误（如不可导出的字段）
	}

	fmt.Println(string(jsonBytes))
	// 输出: {"name":"Alice","age":30}

	return jsonBytes, nil
}

func copeMetric(jManagementMessage *JManagementMessage, metric pmetric.Metric) {
	// 确保 MemoryPool 和 GarbageCollector 的 maps 被初始化
	if jManagementMessage.MemoryPool.MemoryUsages == nil {
		jManagementMessage.MemoryPool.MemoryUsages = make(map[string]*MemoryUsage)
	}
	if jManagementMessage.GarbageCollector.GarbageCollectors == nil {
		jManagementMessage.GarbageCollector.GarbageCollectors = make(map[string]*GarbageCollectorInfo)
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
				if _, exists := jManagementMessage.MemoryPool.MemoryUsages[mappedName]; !exists {
					jManagementMessage.MemoryPool.MemoryUsages[mappedName] = &MemoryUsage{
						Max: -1,
					}
				}
				// 根据字段填充数据
				memoryUsage := jManagementMessage.MemoryPool.MemoryUsages[mappedName]
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
				if _, exists := jManagementMessage.GarbageCollector.GarbageCollectors[garbageCollectorName]; !exists {
					jManagementMessage.GarbageCollector.GarbageCollectors[garbageCollectorName] = &GarbageCollectorInfo{}
				}
				garbageCollectorInfo := jManagementMessage.GarbageCollector.GarbageCollectors[garbageCollectorName]
				garbageCollectorInfo.Name = garbageCollectorName
				garbageCollectorInfo.CollectionCount = dataPoint.Count()
				garbageCollectorInfo.CollectionTime = int(dataPoint.Sum() * 1000)
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
		jManagementMessage.Thread.ThreadCount = threadCount
		jManagementMessage.Thread.PeakThreadCount = threadCount
		jManagementMessage.Thread.DeamonThreadCount = daemonThreadCount
	}
}
