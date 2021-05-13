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
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type Exporter struct {
	Mutex sync.RWMutex
	Metrics []metricInfo
}



type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
	Value float64
	NatsType string
	StreamName string
	ConsumerName string
}

func newStreamMetric(name string, help string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName("nats_jsz", "streams", "total"),
			help,
			[]string{"stream"},
			constLabels,
		),
		Type: t,
		NatsType: "stream",
		StreamName: name,
	}
}
func newConsumerMetric(consumerName string, streamName string, help string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName("nats_jsz", "consumer", "total"),
			help,
			[]string{"consumer"},
			constLabels,
		),
		Type: t,
		NatsType: "consumer",
		StreamName: streamName,
		ConsumerName: consumerName,

	}
}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range e.Metrics {
		ch <- v.Desc
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	e.Scrape(ch)
}


func (e *Exporter) Scrape(ch chan<- prometheus.Metric) {
	for _, v := range e.Metrics {
		switch v.NatsType{
		case "stream":
			v.ScrapeStream(ch)

		case "consumer":
			v.ScrapeConsumer(ch)
		}
	}

}

func (m *metricInfo) ScrapeStream(ch chan<- prometheus.Metric) error {
	stream, err := mgr.LoadStream(m.StreamName)
	if err != nil {
		return err
	}

	info, err := stream.Information()
	if err != nil {
		return err
	}
	ch<- prometheus.MustNewConstMetric(m.Desc, m.Type, float64(info.State.Msgs), m.StreamName)
	return nil
}

func (m *metricInfo) ScrapeConsumer(ch chan<- prometheus.Metric) error {
	conn, err := mgr.LoadConsumer(m.StreamName, m.ConsumerName)
	if err != nil {
		return err
	}

	pending, err := conn.PendingMessages()
	if err != nil {
		return err
	}

	conn.RedeliveryCount()

	ch<- prometheus.MustNewConstMetric(m.Desc, m.Type, float64(pending), m.ConsumerName)

	return nil
}