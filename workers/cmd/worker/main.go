package main

import (
	"context"
	"flag"
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
	inputTopic := flag.String("input-topic", kruntime.String("KINETIX_INPUT_TOPIC", pipeline.DefaultInputTopic), "Kafka input topic")
	outputTopic := flag.String("output-topic", kruntime.String("KINETIX_OUTPUT_TOPIC", pipeline.DefaultOutputTopic), "Kafka output topic")
	group := flag.String("group", kruntime.String("KINETIX_GROUP", "kinetix-worker"), "Kafka consumer group")
	metricsAddr := flag.String("metrics-addr", kruntime.String("KINETIX_METRICS_ADDR", ":8080"), "metrics listen address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics := kruntime.NewMetrics("worker")
	reg := prometheus.NewRegistry()
	metrics.Register(reg)
	server := kruntime.Serve(*metricsAddr, reg)
	defer server.Shutdown(context.Background())

	client, err := kgo.NewClient(
		kgo.SeedBrokers(kruntime.CSVFromValue(*brokers)...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*inputTopic),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		log.Fatalf("create kafka client: %v", err)
	}
	defer client.Close()

	log.Printf("worker consuming %s and producing %s via %s", *inputTopic, *outputTopic, *brokers)
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fetchErr := range errs {
				metrics.Errors.WithLabelValues("fetch").Inc()
				log.Printf("fetch error: %v", fetchErr)
			}
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			start := time.Now()
			metrics.InputRecords.Inc()

			event, err := pipeline.DecodeUserEvent(record.Value)
			if err != nil {
				metrics.Errors.WithLabelValues("decode").Inc()
				log.Printf("decode record at offset %d: %v", record.Offset, err)
				return
			}

			normalized := pipeline.Normalize(event, record.Topic, time.Now())
			value, err := pipeline.EncodeNormalizedEvent(normalized)
			if err != nil {
				metrics.Errors.WithLabelValues("encode").Inc()
				log.Printf("encode normalized event %s: %v", event.ID, err)
				return
			}

			out := &kgo.Record{Topic: *outputTopic, Key: []byte(normalized.ID), Value: value}
			if err := client.ProduceSync(ctx, out).FirstErr(); err != nil {
				metrics.Errors.WithLabelValues("produce").Inc()
				log.Printf("produce normalized event %s: %v", normalized.ID, err)
				return
			}

			if err := client.CommitRecords(ctx, record); err != nil {
				metrics.Errors.WithLabelValues("commit").Inc()
				log.Printf("commit input offset %d: %v", record.Offset, err)
				return
			}

			metrics.OutputRecords.Inc()
			metrics.Latency.Observe(time.Since(start).Seconds())
			log.Printf("processed %s -> %s", event.ID, normalized.ID)
		})
	}
}
