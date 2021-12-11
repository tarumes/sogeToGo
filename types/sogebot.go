package types

type SogeBotPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Engines struct {
		Npm  string `json:"npm"`
		Node string `json:"node"`
	} `json:"engines"`
}
