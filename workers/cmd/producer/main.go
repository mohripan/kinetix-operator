package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/kinetix/kinetix-operator/workers/internal/pipeline"
	kruntime "github.com/kinetix/kinetix-operator/workers/internal/runtime"
)

func main() {
	brokers := flag.String("brokers", kruntime.String("KINETIX_BROKERS", "localhost:9092"), "comma-separated Kafka brokers")
	topic := flag.String("topic", kruntime.String("KINETIX_INPUT_TOPIC", pipeline.DefaultInputTopic), "Kafka input topic")
	interval := flag.Duration("interval", kruntime.Duration("KINETIX_PRODUCE_INTERVAL", 1*time.Second), "produce interval")
	metricsAddr := flag.String("metrics-addr", kruntime.String("KINETIX_METRICS_ADDR", ":8080"), "metrics listen address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics := kruntime.NewMetrics("producer")
	reg := prometheus.NewRegistry()
	metrics.Register(reg)
	server := kruntime.Serve(*metricsAddr, reg)
	defer server.Shutdown(context.Background())

	client, err := kgo.NewClient(kgo.SeedBrokers(kruntime.CSVFromValue(*brokers)...))
	if err != nil {
		log.Fatalf("create kafka client: %v", err)
	}
	defer client.Close()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	log.Printf("producer writing to topic %s via %s", *topic, *brokers)
	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			start := time.Now()
			event := pipeline.NewUserEvent(
				fmt.Sprintf("evt-%06d", i),
				fmt.Sprintf("user-%02d", i%10),
				"page_view",
				now,
				map[string]string{"path": fmt.Sprintf("/demo/%d", i%5), "source": "phase-1"},
			)
			value, err := pipeline.EncodeUserEvent(event)
			if err != nil {
				metrics.Errors.WithLabelValues("encode").Inc()
				log.Printf("encode event: %v", err)
				continue
			}
			record := &kgo.Record{Topic: *topic, Key: []byte(event.ID), Value: value}
			if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
				metrics.Errors.WithLabelValues("produce").Inc()
				log.Printf("produce event %s: %v", event.ID, err)
				continue
			}
			metrics.OutputRecords.Inc()
			metrics.Latency.Observe(time.Since(start).Seconds())
			log.Printf("produced %s", event.ID)
		}
	}
}
