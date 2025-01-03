package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	kafkaHost := getEnvOrDefault("KAFKA_HOST", "195.200.5.15:9092")
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		// User-specific properties that you must set
		"bootstrap.servers": kafkaHost,

		// Fixed properties
		"group.id":          "spectator-consumer",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
		os.Exit(1)
	}

	topic := "spectator-topic"
	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		fmt.Printf("Failed to subscribe consumer to the topic '%s': %s", topic, err)
		os.Exit(1)
	}

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create the streaming wait group
	var wg sync.WaitGroup

	// Set up channel for handling stream events
	eventChan := make(chan string, 1)

	// start the streaming routine
	go videoStreamRoutine(eventChan, &wg)

	// Process messages
	run := true
	fmt.Printf("%s Spectator consumer started\n", time.Now().Format(time.RFC3339))
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Errors are informational and automatically handled by the consumer
				continue
			}

			value := string(ev.Value)
			fmt.Printf("%s Received new event with value '%s' and headers '%v'\n", time.Now().Format(time.RFC3339), value, ev.Headers)
			eventChan <- value
		}
	}

	c.Close()
	eventChan <- "stop"
	wg.Wait()
	close(eventChan)
	fmt.Printf("Spectator terminated.\n")
}

func videoStreamRoutine(eventChannel chan string, wg *sync.WaitGroup) {
	cmd := exec.Command(
		"ffmpeg", "-re", "-stream_loop", "-1", "-i",
		"rtsp://192.168.1.25:554/user=admin&password=&channel=1&stream=0.sdp",
		"-c", "copy", "-f", "rtsp", "-rtsp_transport", "tcp",
		"rtsps://gabriel:7jp73b123@195.200.5.15:8322/garage",
	)
	enabled := false
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	for {
		value := <-eventChannel

		switch value {
		case "start":
			if enabled {
				fmt.Printf("%s streaming is already running\n", time.Now().Format(time.RFC3339))
			}

			if err := cmd.Start(); err != nil {
				fmt.Printf("%s ERROR failed to execute ffmpeg stream: %s\n", time.Now().Format(time.RFC3339), err.Error())

				cmd = exec.Command(
					"ffmpeg", "-re", "-stream_loop", "-1", "-i",
					"rtsp://192.168.1.25:554/user=admin&password=&channel=1&stream=0.sdp",
					"-c", "copy", "-f", "rtsp", "-rtsp_transport", "tcp",
					"rtsps://gabriel:7jp73b123@195.200.5.15:8322/garage",
				)

				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				enabled = false
			} else {
				wg.Add(1)
				enabled = true
				fmt.Printf("%s Video stream started\n", time.Now().Format(time.RFC3339))

				go func() {
					err := cmd.Wait()
					if err != nil {
						fmt.Printf("%s ffmpeg process: %s\n", time.Now().Format(time.RFC3339), err)

						cmd = exec.Command(
							"ffmpeg", "-re", "-stream_loop", "-1", "-i",
							"rtsp://192.168.1.25:554/user=admin&password=&channel=1&stream=0.sdp",
							"-c", "copy", "-f", "rtsp", "-rtsp_transport", "tcp",
							"rtsps://gabriel:7jp73b123@195.200.5.15:8322/garage",
						)

						cmd.Stderr = os.Stderr
						cmd.Stdout = os.Stdout
					}
				}()
			}
		case "stop":
			if !enabled {
				fmt.Printf("%s Currently there is no stream to stop\n", time.Now().Format(time.RFC3339))
				continue
			}

			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("%s ERROR failed to cancel ffmpeg stream: %s\n", time.Now().Format(time.RFC3339), err.Error())
			} else {
				fmt.Printf("%s Video stream stoped\n", time.Now().Format(time.RFC3339))
			}

			enabled = false
			wg.Done()
		default:
			fmt.Printf("%s Received unknown message: %s\n", time.Now().Format(time.RFC3339), value)
		}
	}
}

func getEnvOrDefault(env, defaultValue string) string {
	value, isPresent := os.LookupEnv(env)

	if !isPresent {
		return defaultValue
	}

	return value
}
