package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/ReinforceZwei/network-mon/config"
	"github.com/ReinforceZwei/network-mon/pingwrap"
	"github.com/ReinforceZwei/network-mon/ssh"
)

func main() {
	c, err := config.LoadOrCreateDefault()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Loaded config file from default location")

	var p pingwrap.PingWrap
	if runtime.GOOS == "linux" {
		p = pingwrap.PingLinux{}
	} else if runtime.GOOS == "windows" {
		p = pingwrap.PingWindows{}
	} else {
		log.Fatalf("Platform not supported: %s\n", runtime.GOOS)
	}

	log.Println("Monitoring started")
	for {
		if p.PingOnce(c.TestTarget[0]) {
			// Ping success. Wait for next test cycle
			log.Printf("ping %s success\n", c.TestTarget[0])
		} else {
			// Ping fail. Retry other test target
			log.Printf("ping %s failed. Enter rapid ping mode...\n", c.TestTarget[0])
			pingOk := rapidPingTest(c.MaxFailAttempt, c.FailingIntervalSecond, c.TestTarget, p)
			if pingOk {
				// Resume normal
			} else {
				// Still not success. Restart network on router
				log.Printf("Maximum retry reached. Attempt to restart network...\n")
				conn, err := ssh.Connect(c.Router+":"+"22", c.SshUser, c.SshPassword)
				if err != nil {
					log.Printf("Cannot connect to router: %s\n", err.Error())
				} else {
					pingOk = false
					for i := 0; i < c.RestartAttempt; i++ {
						err = conn.Execute(c.NetworkRestartCommand)
						if err != nil {
							log.Printf("Cannot execute network restart command: %s\n", err.Error())
						} else {
							log.Printf("Restarting. Waiting %d seconds...", c.RestartMaxWaitTime)
							time.Sleep(time.Duration(c.RestartMaxWaitTime) * time.Second)
							if rapidPingTest(c.MaxFailAttempt, c.FailingIntervalSecond, c.TestTarget, p) {
								// Resume normal
								notifyResumeNormal(c.Notify.Url, c.Notify.MessagePayload, "[netmon] Network went out and resumed after network restart")
								pingOk = true
								break
							}
						}
					}
					if !pingOk {
						// Still not success. Reboot router
						// Ignore any error from reboot command
						conn.Execute(c.RouterRebootCommand)
						log.Printf("Restart network doesn't work. Rebooting. Waiting %d seconds...", c.RebootMaxWaitTime)
						if rapidPingTest(c.MaxFailAttempt, c.FailingIntervalSecond, c.TestTarget, p) {
							// Resume normal
							notifyResumeNormal(c.Notify.Url, c.Notify.MessagePayload, "[netmon] Network went out and resumed after reboot")
						} else {
							// Nothing we can do now :(
							log.Println("Unable to resume network. We will keep monitoring the network")
							resumed := false
							for {
								for i := 0; i < len(c.TestTarget); i++ {
									if p.PingOnce(c.TestTarget[i]) {
										resumed = true
										break
									}
								}
								if resumed {
									notifyResumeNormal(c.Notify.Url, c.Notify.MessagePayload, "[netmon] Network went out and resumed itself")
									break
								}
								time.Sleep(time.Duration(c.TestIntervalSecond) * time.Second)
							}
						}
					}
				}
			}
		}
		log.Printf("Sleeping for %d seconds\n", c.TestIntervalSecond)
		time.Sleep(time.Duration(c.TestIntervalSecond) * time.Second)
	}
}

func rapidPingTest(round, interval int, targets []string, p pingwrap.PingWrap) bool {
	for i := 0; i < round; i++ {
		if p.PingOnce(targets[i%len(targets)]) {
			// Resume normal
			log.Printf("rapidPingTest: ping %s success. Resume normal\n", targets[i%len(targets)])
			return true
		}
		log.Printf("rapidPingTest: ping %s failed. Sleeping for %d seconds before next attempt\n", targets[i%len(targets)], interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return false
}

func notifyResumeNormal(webhookUrl, payload, message string) {
	if webhookUrl != "" && payload != "" && message != "" {
		client := http.Client{
			Timeout: 15 * time.Second,
		}
		payload = fmt.Sprintf(payload, message)
		resp, err := client.Post(webhookUrl, "application/json", bytes.NewBuffer([]byte(payload)))
		if err != nil {
			log.Println("notifyResumeNormal: http request error: " + err.Error())
		} else {
			log.Println("notifyResumeNormal: message sent")
			defer resp.Body.Close()
		}
	}
}
