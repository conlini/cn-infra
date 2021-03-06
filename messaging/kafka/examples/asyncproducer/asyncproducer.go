// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/messaging/kafka/examples/utils"
	"strings"
)

var (
	brokerList  = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster. You can also set the KAFKA_PEERS environment variable")
	partitioner = flag.String("partitioner", "hash", "The partitioning scheme to use. Can be `hash`, `manual`, or `random`")
	partition   = flag.Int("partition", -1, "The partition to produce to.")
	debug       = flag.Bool("debug", false, "turns on debug logging")
	silent      = flag.Bool("silent", false, "Turn off printing the message's topic, partition, and offset to stdout")
)

func main() {
	flag.Parse()

	if *brokerList == "" {
		printUsageErrorAndExit("no -brokers specified. Alternatively, set the KAFKA_PEERS environment variable")
	}

	succCh := make(chan *client.ProducerMessage)
	errCh := make(chan *client.ProducerError)

	// init config
	config := client.NewConfig()
	config.SetDebug(*debug)
	config.SetPartition(int32(*partition))
	config.SetPartitioner(*partitioner)
	config.SetSendSuccess(true)
	config.SetSuccessChan(succCh)
	config.SetSendError(true)
	config.SetErrorChan(errCh)
	config.SetBrokers(strings.Split(*brokerList, ",")...)

	// init producer
	producer, err := client.NewAsyncProducer(config, nil)
	if err != nil {
		os.Exit(1)
	}

	go func() {
	eventLoop:
		for {
			select {
			case <-producer.GetCloseChannel():
				break eventLoop
			case msg := <-succCh:
				fmt.Println("message sent successfully - ", msg)
			case err := <-errCh:
				fmt.Println("message errored - ", err)
			}
		}
	}()

	// get command
	for {
		command := utils.GetCommand()
		switch command.Cmd {
		case "quit":
			err := closeProducer(producer)
			if err != nil {
				fmt.Println("terminated abnormally")
				os.Exit(1)
			}
			fmt.Println("ended successfully")
			os.Exit(0)
		case "message":
			err := sendMessage(producer, command.Message)
			if err != nil {
				fmt.Printf("send message error: %v\n", err)
			}

		default:
			fmt.Println("invalid command")
		}
	}
}

// send message
func sendMessage(producer *client.AsyncProducer, msg utils.Message) error {
	var (
		msgKey   []byte
		msgMeta  []byte
		msgValue []byte
	)

	// init message
	if msg.Key != "" {
		msgKey = []byte(msg.Key)
	}
	if msg.Metadata != "" {
		msgMeta = []byte(msg.Metadata)
	}
	msgValue = []byte(msg.Text)

	// send message
	producer.SendMsgByte(msg.Topic, msgKey, msgValue, msgMeta)

	fmt.Println("message sent")
	return nil
}

func closeProducer(producer *client.AsyncProducer) error {
	// close producer
	fmt.Println("closing producer ...")
	err := producer.Close(true)
	if err != nil {
		fmt.Printf("AsyncProducer close errored: %v\n", err)
		return err
	}
	return nil
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	os.Exit(code)
}

func printUsageErrorAndExit(message string) {
	fmt.Fprintln(os.Stderr, "ERROR:", message)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
