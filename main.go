package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-version"
)

func main() {
	fmt.Println("[i] This tool is currently experimental\n keep in mind to make propper backups of your sogebot.db and/or .env file")
	var doInstall bool = false

	getRateLimit()
	info := Collector()

	if info.Install.BotExist {
		vBot, err := version.NewSemver(info.Versions.Bot)
		if err != nil {
			log.Fatal(err)
		}
		vRelease, err := version.NewSemver(info.Versions.Release)
		if err != nil {
			log.Fatal(err)
		}
		if vRelease.GreaterThan(vBot) {
			doInstall = true
		}
	} else {
		// bot dont exist
		doInstall = true
	}

	fmt.Println("BotExist", info.Install.BotExist)
	fmt.Println("doInstall", doInstall)
	fmt.Printf("Bot: '%s'\nRelease: '%s'\n", info.Versions.Bot, info.Versions.Release)

	if doInstall {
		err := os.MkdirAll("./bot", 0777)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll("./backup", 0777)
		if err != nil {
			log.Fatal(err)
		}
		if info.Install.BotExist {
			_ = os.Mkdir("./temp", 0777)
			envFile, _ := exists("./bot/.env")
			dbFile, _ := exists("./bot/sogebot.db")
			if envFile {
				fmt.Println("[i] move previous config to temp dir")
				err = copyFile("./bot/.env", "./temp/.env")
				if err != nil {
					log.Fatal(".env copy failed", err)
				} else {
					log.Println("\".env\" copy done")
				}
			}
			if dbFile {
				fmt.Println("[i] move previous database to temp dir")
				err = copyFile("./bot/sogebot.db", fmt.Sprintf("./backup/sogebot_%d.db", time.Now().Unix()))
				if err != nil {
					log.Fatal("backup failed", err)
				}
				err = copyFile("./bot/sogebot.db", "./temp/sogebot.db")
				if err != nil {
					log.Fatal(err)
				} else {
					log.Println("\"sogebot.db\" copy done")
				}
			}
			err = os.RemoveAll("./bot/")
			if err != nil {
				log.Fatal(err)
			} else {
				log.Println("cleanup done")
			}
			err = os.Mkdir("./bot", 0777)
			log.Println(err)
		}

		err = deployBot(info.Git.DownloadURL, info.Git.ZipName)
		if err != nil {
			log.Fatal(err)
		}

		if info.Install.BotExist {
			envFile, _ := exists("./temp/.env")
			dbFile, _ := exists("./temp/sogebot.db")
			if envFile {
				log.Println("[i] move previous config to bot dir")
				err = copyFile("./temp/.env", "./bot/.env")
				if err != nil {
					log.Fatal("no .env found from previous install")
				} else {
					log.Println("copy done")
				}
			}
			if dbFile {
				log.Println("[i] move previous database to bot dir")
				err = copyFile("./temp/sogebot.db", "./bot/sogebot.db")
				if err != nil {
					log.Fatal("no sogebot.db found from previous install")
				} else {
					log.Println("copy done")
				}
			}

			log.Println("[i] cleanup artifacts")
			err = os.RemoveAll("./temp")
			if err != nil {
				log.Fatal("delete temp folder failed ", err)
			} else {
				log.Println("cleanup done")
			}
		}

		log.Println("start install => this may take a while")
		_, err = exe("./bot", "npm", []string{"install"}, true)
		if err != nil {
			log.Fatal(err)
		}
	}

	exe("./bot", "npm", []string{"start"}, true)

}
