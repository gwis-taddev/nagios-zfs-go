package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	VERSION = "0.0.2"
)

type zpool struct {
	name     string
	capacity int64
	healthy  bool
	faulted  int64
}

func checkHealth(z *zpool, output string) (err error) {
	output = strings.Trim(output, "\n")
	if output == "ONLINE" {
		z.healthy = true
	} else if output == "DEGRADED" || output == "FAULTED" {
		z.healthy = false
	} else {
		z.healthy = false // just to make sure
		err = errors.New("Unknown status")
	}
	return err
}

func getCapacity(z *zpool, output string) (err error) {
	s := strings.Split(output, "%")[0]
	z.capacity, err = strconv.ParseInt(s, 0, 8)
	if err != nil {
		return err
	}
	return err
}

func getFaulted(z *zpool, output string) (err error) {
	lines := strings.Split(output, "\n")
	status := strings.Split(lines[1], " ")[2]
	if status == "ONLINE" {
		z.faulted = 0 // assume ONLINE means no faulted/unavailable providers
	} else if status == "DEGRADED" {
		var count int64
		for _, line := range lines {
			if strings.Contains(line, "FAULTED") || strings.Contains(line, "UNAVAIL") {
				count = count + 1
			}
		}
		z.faulted = count
	} else {
		z.faulted = 1 // fake faulted if there is a parsing error
		err = errors.New("Error parsing faulted/unavailable disks")
	}
	return
}

func runZpoolCommand(args []string) string {
	zpool_path, err := exec.LookPath("zpool")
	if err != nil {
		log.Fatal("Could not find zpool in PATH")
	}
	cmd := exec.Command(zpool_path, args...)
	out, _ := cmd.CombinedOutput()
	return fmt.Sprintf("%s", out)
}

func main() {
	z := zpool{name: "tank"}
	output := runZpoolCommand([]string{"status", z.name})
	err := getFaulted(&z, output)
	if err != nil {
		log.Fatal("Error parsing zpool status")
	}

	output = runZpoolCommand([]string{"list", "-H", "-o", "health", z.name})
	err = checkHealth(&z, output)
	if err != nil {
		log.Fatal("Error parsing zpool list -H -o health ", z.name)
	}

	output = runZpoolCommand([]string{"list", "-H", "-o", "cap", z.name})
	err = getCapacity(&z, output)
	if err != nil {
		log.Fatal("Error parsing zpool capacity")
	}
	fmt.Println(z)
}