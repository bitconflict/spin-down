package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type State struct {
	Disk string
	IO   int
}

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
	if err != nil {
		fmt.Println("Error opening diskstats")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLine := scanner.Text()
		str_arr := strings.Split(currentLine, " ")
		disk_str := str_arr[11]
		if disk_str == disk {
			iotime, err = strconv.Atoi(str_arr[22])
			break
		}
	}
	return iotime, err
}

func getPreviousIOTime(disk string, state *State) (iotime int, err error) {
	// text_buffer := make([]byte, 6)
	content, err := os.ReadFile("/tmp/spin-down-data.json")
	if err != nil {
		fmt.Printf("Failed to read state file with error %v", iotime)
	}
	// Unmarshal the dataiotime = state.IO
	err = json.Unmarshal(content, &state)
	if err != nil {
		fmt.Printf("Error reading state file with error: %v \n", err)
	}
	fmt.Printf("Current disk is %s , prev disk is %s \n", disk, state.Disk)
	if disk == state.Disk {
		iotime = state.IO
		err = nil
	} else {
		err = fmt.Errorf("failed to find state for disk %s", disk)
	}
	return
}

func writeCurrentIOTime(state *State) (err error) {
	content, err := json.Marshal(state)
	if err != nil {
		fmt.Println("Error marshaling data")
	}
	os.WriteFile("/tmp/spin-down-data.json", content, 0777)
	return
}

func isDiskRunningForNoReason(disk string) (beingWasteful bool, err error) {
	beingWasteful = false
	err = nil
	var currentIOTime int
	currentIOTime, err = getDiskIOTime(disk)
	if err != nil {
		fmt.Println("Couldn't read diskIOTime")
		return
	}
	var previousState State
	var currentState State
	var previousIOTime int
	previousIOTime, err = getPreviousIOTime(disk, &previousState)
	if err != nil {
		fmt.Printf("Ran into an error getting previous IO Time with error: %s \n", err)
		currentState.Disk = disk
		currentState.IO = currentIOTime
		disk_err := writeCurrentIOTime(&currentState)
		if disk_err != nil {
			fmt.Printf("Failed to write file with error %v", err)
		}
		return
	}
	fmt.Printf("Values of both IO times are %d and %d \n", currentIOTime, previousIOTime)
	if currentIOTime == previousIOTime && currentIOTime != 0 {
		beingWasteful = true
	} else {
		currentState.Disk = disk
		currentState.IO = currentIOTime
		err = writeCurrentIOTime(&currentState)
		if err != nil {
			fmt.Printf("Failed to write file with error %v", err)
		}
	}
	return
}

func spinDiskDown(disk string) (err error) {
	cmd := exec.Command("hdparm", "-y", "/dev/"+disk)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running hdparm command")
	}
	fmt.Println(string(output[:]))
	return
}

func isDiskCurrentlySpinning(disk string) (spinning bool, err error) {
	spinning = true
	cmd := exec.Command("hdparm", "-C", "/dev/"+disk)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running hdparm status command")
	}
	string_output := string(output[:])
	fmt.Println(string_output)
	if strings.Contains(string_output, "standby") {
		spinning = false
	}
	return
}

func main() {
	disk, timeout := parseArgs()
	fmt.Printf("Selected disk is %v and setting timeout to be %v \n", disk, timeout)
	isDiskOn, err := isDiskCurrentlySpinning(disk)
	if err != nil {
		fmt.Printf("Failed to check current disk status with error: %v", err)
	}
	if !isDiskOn {
		fmt.Println("Disk is already inactive. Exiting")
		return
	}
	beingWasteful, err := isDiskRunningForNoReason(disk)
	if err != nil {
		fmt.Printf("Ran into an error checking if disk is running for no reason with error: %s \n", err)
		return
	}
	if beingWasteful {
		err := spinDiskDown(disk)
		if err != nil {
			fmt.Println("Problem trying to spin disk down.")
		}
		fmt.Println("Disk spun down.")
	} else {
		fmt.Println("Not being wasteful. No action taken.")
	}
}
