package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
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

	handleTestArgs(os.Args[1:], c)

	var p pingwrap.PingWrap
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
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
								conn.Close()
								break
							}
						}
					}
					if !pingOk {
						// Still not success. Reboot router
						log.Printf("Restart network doesn't work")
						resumed := false
						for i := 0; i < c.RebootAttempt; i++ {
							// Ignore any error from reboot command
							conn.Execute(c.RouterRebootCommand)
							log.Printf("Rebooting. Waiting %d seconds...", c.RebootMaxWaitTime)
							time.Sleep(time.Duration(c.RebootMaxWaitTime) * time.Second)
							if rapidPingTest(c.MaxFailAttempt, c.FailingIntervalSecond, c.TestTarget, p) {
								// Resume normal
								notifyResumeNormal(c.Notify.Url, c.Notify.MessagePayload, "[netmon] Network went out and resumed after reboot")
								conn.Close()
								resumed = true
								break
							}
						}
						if !resumed {
							// Nothing we can do now :(
							log.Println("Unable to resume network. We will keep monitoring the network")
							conn.Close()
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
		//log.Printf("Sleeping for %d seconds\n", c.TestIntervalSecond)
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
		payload = fmt.Sprintf(payload, message)
		err := httpPost(webhookUrl, "application/json", payload)
		if err != nil {
			log.Println("notifyResumeNormal: http request error: " + err.Error())
		} else {
			log.Println("notifyResumeNormal: message sent")
		}
	}
}

func httpPost(url, contentType, body string) error {
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Post(url, contentType, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	} else {
		defer resp.Body.Close()
		return nil
	}
}

func handleTestArgs(args []string, c *config.AppConfig) {
	if len(args) > 0 {
		if strings.ToLower(args[0]) == "test" {
			if len(args) > 1 {
				switch strings.ToLower(args[1]) {
				case "webhook":
					testWebhook(c)

				case "ping":
					testPing(c)

				case "ssh":
					testSsh(c)

				case "all":
					testWebhook(c)
					testPing(c)
					testSsh(c)

				default:
					log.Println("Unknown test function")
				}
			} else {
				log.Println("Test function. Available function: webhook, ping, ssh, all")
			}
			os.Exit(0)
		}
	}
}

func testWebhook(c *config.AppConfig) {
	payload := fmt.Sprintf(c.Notify.MessagePayload, "[netmon] Test message")
	err := httpPost(c.Notify.Url, "application/json", payload)
	if err != nil {
		log.Println("webhook test: http request error: " + err.Error())
	} else {
		log.Println("webhook test: ok")
	}
}

func testSsh(c *config.AppConfig) {
	conn, err := ssh.Connect(c.Router+":"+"22", c.SshUser, c.SshPassword)
	if err == nil {
		err = conn.Execute("pwd")
		if err == nil {
			log.Println("ssh test: ok")
			return
		}
	}
	log.Println("ssh test: error: " + err.Error())
}

func testPing(c *config.AppConfig) {
	var p pingwrap.PingWrap
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		p = pingwrap.PingLinux{}
	} else if runtime.GOOS == "windows" {
		p = pingwrap.PingWindows{}
	} else {
		log.Fatalf("Platform not supported: %s\n", runtime.GOOS)
	}
	for _, ip := range c.TestTarget {
		if p.PingOnce(ip) {
			log.Printf("ping test: %s ok\n", ip)
		} else {
			log.Printf("ping test: %s failed\n", ip)
		}
	}
}
