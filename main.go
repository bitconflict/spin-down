package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

func getDiskIOTime(disk string) (iotime string, err error) {
	file, err := os.Open("/proc/diskstats")
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			currentLine := scanner.Text()
			str_arr := strings.Split(currentLine, " ")
			disk_str := str_arr[11]
			if disk_str == disk {
				iotime = str_arr[22]
				break
			}
		}
	}
	defer file.Close()
	return iotime, err
}
func getPreviousIOTime(disk string) (iotime string) {
	file, err := os.Open("/tmp/spin-down-data")
	if err != nil {
		fmt.Printf("failed to open data file with error: %v", err)
	}
	defer file.Close()
	text_buffer := make([]byte, 6)
	_, err = file.Read(text_buffer)
	if err != nil {
		fmt.Printf("Failed to read file with error %v", iotime)
	}
	// text_buffer = []byte(strings.TrimSpace())
	// res := bytes.TrimSpace(text_buffer)
	iotime = string(text_buffer)
	// iotime = strings.TrimSpace(iotime)
	return
}

func writeCurrentIOTime(disk string, currentIOTime string) (err error) {
	os.WriteFile("/tmp/spin-down-data", []byte(currentIOTime), 0777)
	return
}

func isDiskRunningForNoReason(disk string) (beingWasteful bool) {
	beingWasteful = false
	currentIOTime, err := getDiskIOTime(disk)
	if err != nil {
		fmt.Println("Couldn't read diskIOTime")
	}
	fmt.Printf("current io %s \n", currentIOTime)
	previousIOTime := getPreviousIOTime(disk)
	fmt.Printf("previous io %s \n", previousIOTime)
	fmt.Printf("Lengths of both strings are %d and %d \n", len(currentIOTime), len(previousIOTime))
	if currentIOTime == previousIOTime {
		beingWasteful = true
	} else {
		err = writeCurrentIOTime(disk, fmt.Sprint(currentIOTime))
		if err != nil {
			fmt.Printf("Failed to write file with error %v", err)
		}
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
	beingWasteful := isDiskRunningForNoReason(disk)
	if beingWasteful {
		err := spinDiskDown(disk)
		if err != nil {
			fmt.Println("Problem trying to spin disk down.")
		}
		fmt.Println("Disk spun down.")
	} else {
		fmt.Println("Disk is active. No action taken.")
	}
}
