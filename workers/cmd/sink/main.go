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
	topic := flag.String("topic", kruntime.String("KINETIX_OUTPUT_TOPIC", pipeline.DefaultOutputTopic), "Kafka output topic")
	group := flag.String("group", kruntime.String("KINETIX_GROUP", "kinetix-sink"), "Kafka consumer group")
	count := flag.Int("count", kruntime.Int("KINETIX_EXPECTED_COUNT", 0), "records to read before exiting; 0 tails forever")
	timeout := flag.Duration("timeout", kruntime.Duration("KINETIX_TIMEOUT", 2*time.Minute), "maximum wait time")
	metricsAddr := flag.String("metrics-addr", kruntime.String("KINETIX_METRICS_ADDR", ":8080"), "metrics listen address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *count > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	metrics := kruntime.NewMetrics("sink")
	reg := prometheus.NewRegistry()
	metrics.Register(reg)
	server := kruntime.Serve(*metricsAddr, reg)
	defer server.Shutdown(context.Background())

	client, err := kgo.NewClient(
		kgo.SeedBrokers(kruntime.CSVFromValue(*brokers)...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
	)
	if err != nil {
		log.Fatalf("create kafka client: %v", err)
	}
	defer client.Close()

	log.Printf("sink reading %s via %s", *topic, *brokers)
	seen := 0
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			if *count > 0 && seen < *count {
				log.Fatalf("timed out after reading %d/%d records: %v", seen, *count, ctx.Err())
			}
			return
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fetchErr := range errs {
				metrics.Errors.WithLabelValues("fetch").Inc()
				log.Printf("fetch error: %v", fetchErr)
			}
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			metrics.InputRecords.Inc()
			seen++
			log.Printf("sink received key=%s value=%s", string(record.Key), string(record.Value))
			if *count > 0 && seen >= *count {
				stop()
				break
			}
		}
	}
}
