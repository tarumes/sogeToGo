package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func run(dir string, name string, arg []string, printlive bool) ([]string, error) {
	os.Chdir(dir)
	cmd := exec.Command(name, arg...)
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	progress := make(chan struct{})
	scanner := bufio.NewScanner(r)
	var result []string = make([]string, 0)
	go func() {
		// Read line by line and process it
		for scanner.Scan() {
			line := bytes.NewBufferString(scanner.Text())
			result = append(result, line.String())
			if printlive {
				fmt.Println(line)
			}

		}
		// We're all done, unblock the channel
		progress <- struct{}{}
	}()
	cmd.Start()
	<-progress
	err := cmd.Wait()
	if err != nil {
		return result, err
	}
	return result, nil
}
