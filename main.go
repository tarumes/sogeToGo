package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/artdarek/go-unzip/pkg/unzip"
	"github.com/tidwall/gjson"
)

func main() {
	fmt.Println("[i] This tool is currently experimental\n keep in mind to make propper backups of your sogebot.db and/or .env file")
	fmt.Println(" Press ENTER to continue ...")
	awaitUserIput()

	//check if NodeJS is installed
	fmt.Println("Check if NodeJS is installed")
	path, err := exec.LookPath("node")
	if err != nil {
		fmt.Println("✕ No NodeJS installation found")
		fmt.Println("Please checkout official sogeBot Documentation for help")
		os.Exit(1)
	}
	fmt.Println("✓ Found NodeJS in", path)

	/*
		Check if bot is already installed
	*/
	packageJson, err := ioutil.ReadFile("bot/package.json")
	if err != nil {
		fmt.Println("[i] No bot install found")
		fmt.Println("[i] i install that for you")
	} else {
		fmt.Println("[i] Bot install found")
	}
	currentVersion := gjson.Get(string(packageJson), "version")
	fmt.Println(" Current Version:", currentVersion)

	/*
		Check Current Version
	*/
	jsonData, err := getContent("https://api.github.com/repos/sogehige/sogeBot/releases/latest")
	if err != nil {
		panic(err)
	}
	latestZip := gjson.Get(jsonData, "assets.0.browser_download_url")
	latestVersion := gjson.Get(jsonData, "tag_name")
	fmt.Println(" Latest Version is:", latestVersion)
	zipName := gjson.Get(jsonData, "assets.0.name")

	if currentVersion.String() == latestVersion.String() {
		fmt.Println("✓ Your bot is up to date")
	} else {
		fmt.Println("[i] new bot version found\n\nStarting Update")
		// Check for temp folder
		fmt.Println("[i] This tool is currently experimental\n keep in mind to make propper backups of your sogebot.db and/or .env file")
		fmt.Println(" Press CTRL + C to abort")
		fmt.Println(" Press ENTER to continue ...")
		awaitUserIput()

		err = os.Mkdir("./temp", 0777)
		if err != nil {
			fmt.Println("✕ No Permission to write in this directory")
			os.Exit(1)
		}

		err = copyFile("./bot/sogebot.db", "./temp/sogebot.db")
		if err != nil {
			fmt.Println("[*] no sogebot.db found from previous install")
		} else {
			err = copyFile("./bot/.env", "./temp/.env")
			if err != nil {
				fmt.Println("[*] no .env found from previous install")
			}
		}

		err = os.RemoveAll("./bot/")
		if err != nil {
			fmt.Println("✕ Delete bot folder failed")
			log.Fatal(err)
			os.Exit(1)
		}
		_ = os.Mkdir("./bot", 0777)

		fileUrl := latestZip.String()

		fmt.Println("Try Downloading from\n ", latestZip.String())

		err = DownloadFile("./temp/"+zipName.String(), latestZip.String())
		if err != nil {
			fmt.Println("✕ Download Failed")
		}
		fmt.Println("✓ Downloaded:\n " + fileUrl)

		uz := unzip.New()
		fmt.Println("Unzip:", zipName)
		fmt.Println(" This take a moment")
		files, err := uz.Extract("./temp/"+zipName.String(), "./bot")
		if err != nil {
			fmt.Println("✕ Extracting Failed")
		}
		//fmt.Printf("files list: %v\n", files)
		fmt.Printf(" Extracted files count: %d\n", len(files))

		fmt.Println("✓ Extract Done")
		fmt.Println("[i] starting install\n this may take a while get some coffee while im installing")
		err = copyFile("./bot/sogebot.db", "./temp/sogebot.db")
		if err != nil {
			fmt.Println("[*] no sogebot.db found from previous install")
		} else {
			err = copyFile("./bot/.env", "./temp/.env")
			if err != nil {
				fmt.Println("[*] no .env found from previous install")
			}
		}
		err = os.RemoveAll("./temp")
		if err != nil {
			fmt.Println("✕ Delete temp folder failed")
			log.Fatal(err)
			os.Exit(1)
		}

		os.Chdir("./bot")
		cmd := exec.Command("npm", "ci")
		err = cmd.Run()
		if err != nil {
			fmt.Println("✕ Install Failed")
			log.Fatal(err)
		}
		fmt.Println("✓ Your bot is installed and up to date\ngo into the new bot folder and run `npm start`\nenjoy sogeBot")
	}
	fmt.Println(" Press ENTER to continue ...")
	awaitUserIput()

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

func awaitUserIput() {
	buf := bufio.NewReader(os.Stdin)
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(sentence))
	}
}
