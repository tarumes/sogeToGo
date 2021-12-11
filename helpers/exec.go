package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// run commandline when printlive is true return slice is empty
func Exe(dir string, name string, arg []string, printlive bool) ([]string, error) {
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
			if printlive {
				fmt.Println(line)
			} else {
				result = append(result, line.String())
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
