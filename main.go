package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sogeToGo/helpers"
	"sogeToGo/types"
	"strings"
	"time"
)

func main() {
	//var doInstall bool = false
	//var doStart bool = false

	var timestamp string = fmt.Sprint(time.Now().Unix())
	err := os.MkdirAll(fmt.Sprintf("backup/%s/", timestamp), 0755)
	if err != nil {
		log.Fatal(err)
	}
	helpers.CopyFile("bot/sogebot.db", fmt.Sprintf("backup/%s/sogebot.db", timestamp))
	helpers.CopyFile("bot/sogebot.db-wal", fmt.Sprintf("backup/%s/sogebot.db-wal", timestamp))
	helpers.CopyFile("bot/sogebot.db-shm", fmt.Sprintf("backup/%s/sogebot.db-shm", timestamp))
	helpers.CopyFile("bot/.env", fmt.Sprintf("backup/%s/.env", timestamp))

	release, err := func() (types.GithubApiRelease, error) {
		data, err := helpers.GetURLContent("https://api.github.com/repos/sogebot/sogeBot/releases/latest")
		if err != nil {
			return types.GithubApiRelease{}, err
		}
		var reply types.GithubApiRelease
		err = json.Unmarshal(data, &reply)
		if err != nil {
			fmt.Println(err)
			return types.GithubApiRelease{}, err
		}
		return reply, err
	}()
	if err != nil {
		fmt.Println(err)
	}

	installed, err := func() (types.SogeBotPackage, error) {
		data, err := os.ReadFile("bot/package.json")
		if err != nil {
			return types.SogeBotPackage{}, err
		}
		var reply types.SogeBotPackage
		err = json.Unmarshal(data, &reply)
		if err != nil {
			fmt.Println(err)
			return types.SogeBotPackage{}, err
		}
		return reply, err
	}()
	if err != nil {
		fmt.Println(err)
		installed.Version = ""
	}

	func() {
		if release.TagName != installed.Version || installed.Version == "" {
			if strings.HasSuffix(release.Assets[0].BrowserDownloadUrl, ".zip") {
				fmt.Println("New Version Found")

				fmt.Println("create temp folder")
				os.MkdirAll("temp", 0755)
				defer os.RemoveAll("temp")

				fmt.Println("download zip version")
				err := helpers.DownloadToFile(release.Assets[0].BrowserDownloadUrl, "temp/sogebot.zip")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("unzip")
				_, err = helpers.Unzip("temp/sogebot.zip", "temp/bot")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("try copy previous files")
				helpers.CopyFile("bot/sogebot.db", "temp/sogebot.db")
				helpers.CopyFile("bot/sogebot.db-wal", "temp/sogebot.db-wal")
				helpers.CopyFile("bot/sogebot.db-shm", "temp/sogebot.db-shm")
				helpers.CopyFile("bot/.env", "temp/.env")

				fmt.Println("clean old bot folder")
				err = os.RemoveAll("bot")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				fmt.Println("create new bot folder")
				err = os.Rename("temp/bot", "bot")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				fmt.Println("try moving back old settings")
				helpers.CopyFile("temp/sogebot.db", "bot/sogebot.db")
				helpers.CopyFile("temp/sogebot.db-wal", "bot/sogebot.db-wal")
				helpers.CopyFile("temp/sogebot.db-shm", "bot/sogebot.db-shm")
				helpers.CopyFile("temp/.env", "bot/.env")

				fmt.Println("install new bot")
				helpers.Exe("bot", "npm", []string{"install"}, true)
			}
		}
	}()

	func() {
		helpers.Exe("bot", "npm", []string{"start"}, true)
	}()
}
