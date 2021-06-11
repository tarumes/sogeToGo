package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/artdarek/go-unzip/pkg/unzip"
	"github.com/tidwall/gjson"
)

func main() {
	fmt.Println("[i] This tool is currently experimental\n keep in mind to make propper backups of your sogebot.db and/or .env file")

	//check for NodeJS
	checkNodeJS()

	// check temp folder
	fmt.Println("[i] check for existing temp folder")
	temp, err := exists("./temp")
	if err != nil {
		fmt.Println(" ✕ Something went wrong with check for temp folder")
	}
	if temp { //bool are checked without operators !!!!
		envFile, _ := exists("./temp/.env")
		if envFile {
			fmt.Println(" ✕ There is an temp directory detected from an previous try\n i refuse to continue until you check these files and delete the folder")
			os.Exit(1)
		}
	} else {
		fmt.Println(" ✓ check ok")
	}

	fmt.Println("[i] check online for new version")
	getRateLimit()
	// check online for latest version
	jsonData, err := getContent("https://api.github.com/repos/sogehige/sogeBot/releases/latest")
	if err != nil {
		fmt.Println(" ✕ Connection to Github failed check for connection issues or firewall settings")
		os.Exit(1)
	}

	latestZip := gjson.Get(jsonData, "assets.0.browser_download_url")
	if latestZip.String() == "" {
		fmt.Println(" ✕ Can`t parse Download Link from Github API")
		os.Exit(1)
	}
	latestVersion := gjson.Get(jsonData, "tag_name")
	if latestVersion.String() == "" {
		fmt.Println(" ✕ Can`t parse latest Version from Github API")
		os.Exit(1)
	}
	fmt.Println(" ✓ check ok")
	zipName := gjson.Get(jsonData, "assets.0.name")

	//check previous Install and make copy of files
	currentVersion, err := checkPreviousInstall()
	if err != nil {
		// install new Bot
		fmt.Println("[i] Found no previous install")
		_ = os.Mkdir("./bot", 0777)
		_ = os.Mkdir("./temp", 0777)
		installBot(latestZip.String(), zipName.String())

		fmt.Println("[i] cleanup artifacts")
		err = os.RemoveAll("./temp")
		if err != nil {
			fmt.Println(" ✕ Delete temp folder failed")
			log.Fatal(err)
			os.Exit(1)
		} else {
			fmt.Println(" ✓ cleanup done")
		}

		fmt.Println("[i] starting install\n this may take a while get some coffee while im installing")

		// run NPM install
		os.Chdir("./bot")
		cmd := exec.Command("npm", "ci")

		r, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		progress := make(chan struct{})
		scanner := bufio.NewScanner(r)
		go func() {

			// Read line by line and process it
			for scanner.Scan() {
				line := bytes.NewBufferString(scanner.Text())
				fmt.Println(line.String())
			}

			// We're all done, unblock the channel
			progress <- struct{}{}

		}()
		cmd.Start()
		<-progress
		err = cmd.Wait()
		fmt.Println(err)

		fmt.Println(" ✓ Your bot is installed and up to date\ngo into the new bot folder and run `npm start`\nenjoy sogeBot")

		cmd = exec.Command("npm", "start")

		r, _ = cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		progress = make(chan struct{})
		scanner = bufio.NewScanner(r)
		go func() {

			// Read line by line and process it
			for scanner.Scan() {
				line := bytes.NewBufferString(scanner.Text())
				fmt.Println(line.String())
			}

			// We're all done, unblock the channel
			progress <- struct{}{}

		}()
		cmd.Start()
		<-progress
		err = cmd.Wait()
		fmt.Println(err)

	} else {
		// update or start current Bot
		if currentVersion == "" {
			fmt.Println(" ✕ parsing from current install failed")
			os.Exit(1)
		} else {
			fmt.Println("[i] Found previous install")
			fmt.Println(" Latest Version is:", latestVersion.String())

			//compare new and old version
			if currentVersion == latestVersion.String() {
				fmt.Println(" ✓ Your bot is up to date")

				_, err := run("./bot", "npm", []string{"start"}, true)
				if err != nil {
					log.Println(err)
				}

			} else {
				fmt.Println("[*] new bot version found\n\nStarting Update")
				_ = os.Mkdir("./temp", 0777)
				envFile, _ := exists("./bot/.env")
				dbFile, _ := exists("./bot/sogebot.db")
				if envFile {
					fmt.Println("[i] move previous config to temp dir")
					err = copyFile("./bot/.env", "./temp/.env")
					if err != nil {
						fmt.Println(" ✕ copy failed", err)
						os.Exit(1)
					} else {
						fmt.Println(" ✓ copy done")
					}
				}
				if dbFile {
					fmt.Println("[i] move previous database to temp dir")
					err = copyFile("./bot/sogebot.db", "./temp/sogebot.db")
					if err != nil {
						fmt.Println(" ✕ copy failed", err)
						os.Exit(1)
					} else {
						fmt.Println(" ✓ copy done")
					}
				}

				fmt.Println("[i] cleanup old bot folder")
				err = os.RemoveAll("./bot/")
				if err != nil {
					fmt.Println(" ✕ Delete bot folder failed")
					log.Fatal(err)
					os.Exit(1)
				} else {
					fmt.Println(" ✓ cleanup done")
				}

				// make new Bot Folder
				_ = os.Mkdir("./bot", 0777)

				// install current version
				installBot(latestZip.String(), zipName.String())

				// copy back old settings

				envFile, _ = exists("./temp/.env")
				dbFile, _ = exists("./temp/sogebot.db")
				if envFile {
					fmt.Println("[i] move previous config to bot dir")
					err = copyFile("./temp/.env", "./bot/.env")
					if err != nil {
						fmt.Println(" ✕ no .env found from previous install")
						os.Exit(1)
					} else {
						fmt.Println(" ✓ copy done")
					}
				}
				if dbFile {
					fmt.Println("[i] move previous database to bot dir")
					err = copyFile("./temp/sogebot.db", "./bot/sogebot.db")
					if err != nil {
						fmt.Println(" ✕ no sogebot.db found from previous install")
					} else {
						fmt.Println(" ✓ copy done")
					}
				}

				fmt.Println("[i] cleanup artifacts")
				err = os.RemoveAll("./temp")
				if err != nil {
					fmt.Println(" ✕ Delete temp folder failed")
					log.Fatal(err)
					os.Exit(1)
				} else {
					fmt.Println(" ✓ cleanup done")
				}

				fmt.Println("[i] starting install\n this may take a while get some coffee while im installing")

				// run NPM install
				_, err := run("./bot", "npm", []string{"ci"}, true)
				if err != nil {
					log.Println(err)
				}

				fmt.Println(" ✓ Your bot is installed and up to date\ngo into the new bot folder and run `npm start`\nenjoy sogeBot")

				_, err = run("./bot", "npm", []string{"start"}, true)
				if err != nil {
					log.Println(err)
				}
			}

		}
	}
	os.Exit(0)
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

/*
func awaitUserIput() {
	buf := bufio.NewReader(os.Stdin)
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(sentence))
	}
}
*/

func checkNodeJS() {
	//check if NodeJS is installed
	fmt.Println("[*] Check if NodeJS is installed")
	path, err := exec.LookPath("node")
	if err != nil {
		fmt.Println("✕ No NodeJS installation found")
		fmt.Println("Please checkout official sogeBot Documentation for help")
		os.Exit(1)
	}
	fmt.Println(" ✓ Found NodeJS in:", path)
}

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

func installBot(latestZip string, zipName string) {
	fileUrl := latestZip

	fmt.Println("Try Downloading from\n ", latestZip)

	err := DownloadFile("./temp/"+zipName, latestZip)
	if err != nil {
		fmt.Println("✕ Download Failed")
	}
	fmt.Println("✓ Downloaded:\n " + fileUrl)

	uz := unzip.New()
	fmt.Println("Unzip:", zipName)
	fmt.Println(" This take a moment")
	files, err := uz.Extract("./temp/"+zipName, "./bot")
	if err != nil {
		fmt.Println("✕ Extracting Failed")
	}
	//fmt.Printf("files list: %v\n", files)
	fmt.Printf(" Extracted files count: %d\n", len(files))

	fmt.Println(" ✓ Extract Done")
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
