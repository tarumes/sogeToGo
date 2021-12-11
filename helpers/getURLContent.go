package helpers

import (
	"io/ioutil"
	"net/http"
)

func GetURLContent(url string) ([]byte, error) {
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

	return html, err
}
