package main

import (
	"flag"
	"fmt"
	"os/exec"
)

func parseArgs() (disk string, timeout int) {
	diskPtr := flag.String("disk", "sda", "Disk name under /dev")
	timeoutPtr := flag.Int("timeout", 600, "Disk inactivity timeout in seconds")
	flag.Parse()
	disk = *diskPtr
	timeout = *timeoutPtr
	return
}

func isDiskRunningForNoReason(disk string, timeout int) (beingWasteful bool) {

	beingWasteful = false
	return
}

func spinDiskDown(disk string) (err error) {
	cmd := exec.Command("ls")
	err = cmd.Run()

	return
}

func main() {

	disk, timeout := parseArgs()
	fmt.Println("Starting")
	beingWasteful := isDiskRunningForNoReason(disk, timeout)
	if beingWasteful {
		spinDiskDown(disk)
	}
	fmt.Println("Done")
}
