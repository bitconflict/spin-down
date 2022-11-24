package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func parseArgs() (disk string, timeout int) {
	diskPtr := flag.String("disk", "sda", "Disk name under /dev")
	timeoutPtr := flag.Int("timeout", 600, "Disk inactivity timeout in seconds")
	flag.Parse()
	disk = *diskPtr
	timeout = *timeoutPtr
	return
}

func getDiskIOTime(disk string) (iotime int, err error) {
	file, err := os.Open("/proc/diskstats")
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			currentLine := scanner.Text()
			str_arr := strings.Split(currentLine, " ")
			disk_str := str_arr[11]
			if disk_str == disk {
				iotime, err = strconv.Atoi(str_arr[22])
			}
		}
	}
	defer file.Close()
	return iotime, err
}
func getPreviousIOTime(disk string) (iotime int) {
	return 0
}

func isDiskRunningForNoReason(disk string, timeout int) (beingWasteful bool) {
	beingWasteful = false
	currentIOTime, err := getDiskIOTime(disk)
	if err != nil {
		fmt.Println("Couldn't read diskIOTime")
	}
	fmt.Printf("current io %v \n", currentIOTime)
	previousIOTime := getPreviousIOTime(disk)
	fmt.Printf("previous io %v \n", previousIOTime)
	if currentIOTime == previousIOTime {
		beingWasteful = true
	}
	return
}

func spinDiskDown(disk string) (err error) {
	// cmd := exec.Command("hdparm", "-y", "/dev/"+disk)
	cmd := exec.Command("echo", "hdparm spun down")
	_, err = cmd.CombinedOutput()
	return
}

func main() {
	disk, timeout := parseArgs()
	fmt.Printf("Selected disk is %v and setting timeout to be %v \n", disk, timeout)
	beingWasteful := isDiskRunningForNoReason(disk, timeout)
	if beingWasteful {
		err := spinDiskDown(disk)
		if err != nil {
			fmt.Println("Problem trying to spin disk down.")
		}
		fmt.Println("Disk spun down.")
	}
	fmt.Println("Done")
}
