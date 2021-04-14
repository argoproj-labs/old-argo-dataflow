package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/tools/tls"
)

var (
	brokerList    = flag.String("brokers", defaultPeers(), "The comma separated list of brokers in the Kafka cluster")
	topic         = flag.String("topic", "input-topic", "The topic to")
	verbose       = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	tlsEnabled    = flag.Bool("tls-enabled", false, "Whether to enable TLS")
	tlsSkipVerify = flag.Bool("tls-skip-verify", false, "Whether skip TLS server cert verification")
	tlsClientCert = flag.String("tls-client-cert", "", "Client cert for client authentication (use with -tls-enabled and -tls-client-key)")
	tlsClientKey  = flag.String("tls-client-key", "", "Client key for client authentication (use with tls-enabled and -tls-client-cert)")

	logger = log.New(os.Stdout, "", log.LstdFlags)
)

func defaultPeers() string {
	if env, ok := os.LookupEnv("KAFKA_PEERS"); ok {
		return env
	}
	return "kafka-0.broker:9092"
}

func main() {
	flag.Parse()
	if *brokerList == "" {
		panic(fmt.Errorf("you have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable"))
	}
	if *topic == "" {
		panic(fmt.Errorf("-topic is required"))
	}
	if *verbose {
		sarama.Logger = logger
	}
	config := sarama.NewConfig()
	if *tlsEnabled {
		tlsConfig, err := tls.NewConfig(*tlsClientCert, *tlsClientKey)
		if err != nil {
			panic(err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Config.InsecureSkipVerify = *tlsSkipVerify
	}

	addrs := strings.Split(*brokerList, ",")
	admin, err := sarama.NewClusterAdmin(addrs, config)
	if err != nil {
		panic(err)
	}
	defer admin.Close()
	producer, err := sarama.NewAsyncProducer(addrs, config)
	if err != nil {
		panic(err)
	}
	defer producer.Close()
	err = func() error {
		cmd := flag.Args()[0]
		switch cmd {
		case "create-topic":
			return createTopicCmd(admin)
		case "pump-topic":
			return pumpTopicCmd(producer)
		default:
			return fmt.Errorf("unknown comand %q", cmd)
		}
	}()
	if err != nil {
		panic(err)
	}
}

var (
	sleep = flag.String("sleep", "1s", "how long to sleep")
)

func pumpTopicCmd(producer sarama.AsyncProducer) error {
	duration, err := time.ParseDuration(*sleep)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "topic %q sleep %v \n", *topic, duration)
	start := time.Now()
	for i := 0; ; i++ {
		x := fmt.Sprintf("%s-%d", FunnyAnimal(), i)
		producer.Input() <- &sarama.ProducerMessage{
			Topic: *topic,
			Value: sarama.StringEncoder(x),
		}
		_, _ = fmt.Fprintf(os.Stdout, "sent %q (%.0f/s)\n", x, (1+float64(i))/time.Since(start).Seconds())
		time.Sleep(duration)
	}
}

func createTopicCmd(admin sarama.ClusterAdmin) error {
	if err := admin.CreateTopic(*topic, &sarama.TopicDetail{NumPartitions: 2, ReplicationFactor: 2}, false); err != nil {
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
