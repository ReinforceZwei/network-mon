package main

import (
	"fmt"

	"github.com/ReinforceZwei/network-mon/config"
	"github.com/ReinforceZwei/network-mon/pingwrap"
)

func main() {
	c, err := config.LoadOrCreateDefault()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n", c)
	p := pingwrap.PingWindows{}
	fmt.Println(p.PingOnce("google.com"))
}
