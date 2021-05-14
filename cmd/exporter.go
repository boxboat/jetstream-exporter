/*
Copyright © 2021 BoxBoat

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

package cmd

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// exporter holds metrics and satisfies the Describe and Collect
// methods required for the Collector interface.
type exporter struct {
	mutex   sync.RWMutex
	metrics []metricInfo
}

// metricInfo holds all of the information about a metric
type metricInfo struct {
	desc         *prometheus.Desc
	promType     prometheus.ValueType
	valueType    string
	value        float64
	natsType     string
	streamName   string
	consumerName string
}

// newStreamMetric returns a metricInfo specific to a stream
func newStreamMetric(name, streamName, help string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nats_jsz", "streams", name),
			help,
			[]string{"stream"},
			constLabels,
		),
		promType:   t,
		valueType:  name,
		natsType:   "stream",
		streamName: streamName,
	}
}

// newConsumerMetric returns a metricInfo specific to a consumer
func newConsumerMetric(name, consumerName, streamName, help string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nats_jsz", "consumer", name),
			help,
			[]string{"consumer"},
			constLabels,
		),
		promType:     t,
		valueType:    name,
		natsType:     "consumer",
		streamName:   streamName,
		consumerName: consumerName,
	}
}

// newExporter returns a new exporter
func newExporter() *exporter {
	return &exporter{}
}

// Describe describes the metrics
func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range e.metrics {
		ch <- v.desc
	}
}

// Collect gets prometheus metrics from the specified servers
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.scrape(ch)
}

// scrape checks the metric type and fetches metrics based on that type
func (e *exporter) scrape(ch chan<- prometheus.Metric) {
	for _, v := range e.metrics {
		switch v.natsType {
		case "stream":
			v.scrapeStream(ch)

		case "consumer":
			v.scrapeConsumer(ch)
		}
	}
}

// scrapeStream fetches the total message and total bytes for a stream
func (m *metricInfo) scrapeStream(ch chan<- prometheus.Metric) error {
	stream, err := mgr.LoadStream(m.streamName)
	if err != nil {
		return err
	}

	info, err := stream.Information()
	if err != nil {
		return err
	}

	switch m.valueType {
	case totalMessage:
		ch <- prometheus.MustNewConstMetric(m.desc, m.promType, float64(info.State.Msgs), m.streamName)
	case totalBytes:
		ch <- prometheus.MustNewConstMetric(m.desc, m.promType, float64(info.State.Bytes), m.streamName)
	}
	return nil
}

// scrapeConsumer fetches the current pending messages and total redelivery count for a consumer
func (m *metricInfo) scrapeConsumer(ch chan<- prometheus.Metric) error {
	conn, err := mgr.LoadConsumer(m.streamName, m.consumerName)
	if err != nil {
		return err
	}

	pending, err := conn.PendingMessages()
	if err != nil {
		return err
	}

	redelivery, err := conn.RedeliveryCount()
	if err != nil {
		return err
	}

	labels := fmt.Sprintf("%s:%s", m.consumerName, m.streamName)

	switch m.valueType {
	case msgPending:
		ch <- prometheus.MustNewConstMetric(m.desc, m.promType, float64(pending), labels)
	case msgRedeliver:
		ch <- prometheus.MustNewConstMetric(m.desc, m.promType, float64(redelivery), labels)
	}

	return nil
}
