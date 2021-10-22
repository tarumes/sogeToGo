package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/artdarek/go-unzip/pkg/unzip"
	"github.com/tidwall/gjson"
)

func cleanVersionString(ver string) string {
	var reply string = ver
	reply = strings.ReplaceAll(reply, "v", "")
	reply = strings.ReplaceAll(reply, "\n", "")
	reply = strings.ReplaceAll(reply, "\r", "")
	reply = strings.ReplaceAll(reply, " ", "")
	reply = strings.ReplaceAll(reply, "<", "")
	reply = strings.ReplaceAll(reply, ">", "")
	reply = strings.ReplaceAll(reply, "=", "")
	reply = strings.ReplaceAll(reply, "^", "")
	return reply
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// run commandline when printlive is true return slice is empty
func exe(dir string, name string, arg []string, printlive bool) ([]string, error) {
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

func getContent(url string) (string, error) {
	//url := "http://tour.golang.org/welcome/1"
	//fmt.Printf("HTML code of %s ...\n", url)
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		panic(err)
	}
	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	// show the HTML code as a string %s
	//fmt.Printf("%s\n", html)
	return string(html), err
}

func getRateLimit() {
	jsonData, err := getContent("https://api.github.com/rate_limit")
	if err != nil {
		fmt.Println(" ✕ Can´t check Github Ratelimit")
		os.Exit(1)
	}
	coreLimit := gjson.Get(jsonData, "resources.core.remaining")
	//coreLimitInt, _ := strconv.ParseInt(coreLimit.String(), 10, 64)
	coreReset := gjson.Get(jsonData, "resources.core.reset")

	i, err := strconv.ParseInt(coreReset.String(), 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	fmt.Println("[i] Github Ratelimit: "+coreLimit.String()+" requests left until", tm)
	if coreLimit.Int() < 10 {
		sleep := (time.Duration(coreReset.Int()) - time.Duration(time.Now().Unix())) * time.Second
		fmt.Println(" ✕ Github Ratelimit reached sleeping for", sleep)
		time.Sleep(sleep)
		fmt.Println(" ✓ Sleep Over.....")
	}

}

func deployBot(latestZip string, zipName string) error {

	temp, _ := exists("temp")
	bot, _ := exists("bot")
	if !temp {
		err := os.Mkdir("./temp", 0777)
		if err != nil {
			return fmt.Errorf("mkdir \"temp\" %s", err)
		}
	}
	if !bot {
		err := os.Mkdir("bot", 0777)
		if err != nil {
			return fmt.Errorf("mkdir \"bot\" %s", err)
		}
	}

	err := DownloadFile("temp/"+zipName, latestZip)
	if err != nil {
		return fmt.Errorf("download %s", err)
	}
	uz := unzip.New()
	_, err = uz.Extract("temp/"+zipName, "./bot")
	if err != nil {
		return fmt.Errorf("unzip %s", err)
	}
	return nil
}

/*
// checkPreviousInstall checks for previous version installed an returns a version string
func checkPreviousInstall() (string, error) {
	packageJson, err := ioutil.ReadFile("bot/package.json")
	if err != nil {
		return "", err
	} else {
		currentVersion := gjson.Get(string(packageJson), "version")
		return currentVersion.String(), nil
	}
}
*/
