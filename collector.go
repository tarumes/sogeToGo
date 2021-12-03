package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/tidwall/gjson"
)

type CollectorData struct {
	Versions struct {
		Node    string
		NPM     string
		Bot     string
		Release string
	}
	Install struct {
		TempCheck bool
		BotExist  bool
	}
	Path struct {
		Node string
	}
	Git struct {
		DownloadURL string
		ZipName     string
	}
}

func Collector() CollectorData {
	var await sync.WaitGroup
	var data CollectorData = CollectorData{}

	//check if NodeJS is installed
	await.Add(1)
	go func() {
		defer await.Done()
		var err error
		data.Path.Node, err = func() (string, error) {
			reply, err := exec.LookPath("node")
			if err != nil {
				return "", err
			}
			return reply, nil
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	//check NodeJS version
	await.Add(1)
	go func() {
		defer await.Done()
		var err error
		data.Versions.Node, err = func() (string, error) {
			reply, err := exec.Command("node", "--version").Output()
			if err != nil {
				return "", err
			}
			node := string(reply)
			return cleanVersionString(node), nil
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	//check NPM Version
	await.Add(1)
	go func() {
		defer await.Done()
		var err error
		data.Versions.NPM, err = func() (string, error) {
			reply, err := exec.Command("npm", "--version").Output()
			if err != nil {
				return "", err
			}
			npm := string(reply)
			return cleanVersionString(npm), nil
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	//check for temp folder
	await.Add(1)
	go func() {
		defer await.Done()
		var err error
		data.Install.TempCheck, err = func() (bool, error) {
			// check temp folder
			fmt.Println("[i] check for existing temp folder")
			temp, err := exists("./temp")
			if err != nil {
				//fmt.Println(" ✕ Something went wrong with check for temp folder")
				return false, err
			}
			if temp { //bool are checked without operators !!!!
				envFile, _ := exists("./temp/.env")
				if envFile {
					return false, fmt.Errorf("files from aborted install found, check and cleanup temp folder")
				}
			}
			return true, nil
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	//check for current bot version
	await.Add(1)
	go func() {
		defer await.Done()
		var err error
		data.Versions.Bot, data.Install.BotExist, err = func() (string, bool, error) {
			// check for current Bot version
			packageJson, err := ioutil.ReadFile("bot/package.json")
			var version BotPackage = BotPackage{}
			json.Unmarshal(packageJson, &version)

			if err != nil {
				return "", false, nil
			} else {
				return version.Version, true, nil
			}
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	//check for current bot release version
	await.Add(1)
	go func() {
		defer await.Done()
		err := func() error {
			// check online for latest version
			jsonData, err := getContent("https://api.github.com/repos/sogebot/sogeBot/releases/latest")
			if err != nil {
				return err
			}

			latestZip := gjson.Get(jsonData, "assets.0.browser_download_url")
			if latestZip.String() == "" {
				return fmt.Errorf("can´t parse JSON from github")
			}
			latestVersion := gjson.Get(jsonData, "tag_name")
			if latestVersion.String() == "" {
				return fmt.Errorf("can´t parse JSON from github")
			}
			zipName := gjson.Get(jsonData, "assets.0.name")
			if zipName.String() == "" {
				return fmt.Errorf("can´t parse JSON from github")
			}
			data.Git.DownloadURL = latestZip.String()
			data.Git.ZipName = zipName.String()
			data.Versions.Release = latestVersion.String()
			return nil
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	await.Wait()
	return data
}
