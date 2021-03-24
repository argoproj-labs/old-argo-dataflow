package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/tools/tls"
)

var (
	brokerList    = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topic         = flag.String("topic", "", "REQUIRED: the topic to consume")
	verbose       = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	tlsEnabled    = flag.Bool("tls-enabled", false, "Whether to enable TLS")
	tlsSkipVerify = flag.Bool("tls-skip-verify", false, "Whether skip TLS server cert verification")
	tlsClientCert = flag.String("tls-client-cert", "", "Client cert for client authentication (use with -tls-enabled and -tls-client-key)")
	tlsClientKey  = flag.String("tls-client-key", "", "Client key for client authentication (use with tls-enabled and -tls-client-cert)")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	if err := mainE(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func mainE() error {
	flag.Parse()

	if *brokerList == "" {
		return fmt.Errorf("you have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	}

	if *topic == "" {
		return fmt.Errorf("-topic is required")
	}

	if *verbose {
		sarama.Logger = logger
	}

	config := sarama.NewConfig()
	if *tlsEnabled {
		tlsConfig, err := tls.NewConfig(*tlsClientCert, *tlsClientKey)
		if err != nil {
			return err
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Config.InsecureSkipVerify = *tlsSkipVerify
	}

	admin, err := sarama.NewClusterAdmin(strings.Split(*brokerList, ","), config)
	if err != nil {
		return err
	}

	if err := admin.CreateTopic(*topic, &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}, false); err != nil {
		if terr, ok := err.(*sarama.TopicError); ok && terr.Err == sarama.ErrTopicAlreadyExists {
			_, _ = fmt.Fprintf(os.Stdout, "topic %q already exists\n", *topic)
			return nil
		} else {
			return err
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "topic %q created\n", *topic)

	return nil
}
