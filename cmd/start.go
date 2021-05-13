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
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type Consumer struct {
	Name string
	Stream string
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the exporter",
	Run: start,
}

var (
	nc = &nats.Conn{}

	mgr = &jsm.Manager{}

	scrape_interval = 15

	prevMessages = 0.0

	streamSlice = []string{"STATUS", "BUILDS"}
	consumers = map[string]string{
		"STATUS": "FAILED",
	}

	streamMetrics = []metricInfo{}
	consumerMetrics = []metricInfo{}

	totalMessage = "total_msg"
	totalBytes = "total_bytes"
	msgPending = "msg_pending"
	msgRedeliver = "msg_redelivery"

)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("servers", "s", []string{"nats://localhost:4222"}, "NATS Servers")
}

func start(cmd *cobra.Command, args []string) {
	var consumers []Consumer
	viper.UnmarshalKey("consumers", &consumers)

	exporter := NewExporter()
	for _, v := range viper.GetStringSlice("streams") {
		msg := newStreamMetric(totalMessage, "Total number of stream messages", prometheus.CounterValue, nil)
		bytes := newStreamMetric(totalBytes, "Total number of stream bytes", prometheus.CounterValue, nil)
		exporter.Metrics = append(exporter.Metrics, msg, bytes)
	}
	for _, v := range consumers {
		msg := newConsumerMetric(msgPending, v.Stream,"Current number of consumer messages", prometheus.GaugeValue, nil)
		redelivery := newConsumerMetric(msgRedeliver, v.Stream, "Number of redelivers", prometheus.CounterValue, nil)
		exporter.Metrics = append(exporter.Metrics, msg)
	}

	prometheus.MustRegister(exporter)

	nc, err := nats.Connect(strings.Join(viper.GetStringSlice("servers"), ","))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	mgr, err = jsm.New(nc)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := fmt.Sprintf("<html>" +
			"<head><title>JSZ Exporter</title></head>" +
			"<body>\n<h1>JSZ Exporter</h1>" +
			"<p><a href='/metrics'>Metrics</a></p>" +
			"</body>\n</html>")
		fmt.Fprint(w, resp)
	})
	log.Fatal(http.ListenAndServe(":9999", nil))

}