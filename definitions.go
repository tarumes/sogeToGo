package main

type BotPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Private     bool     `json:"private"`
	SideEffects []string `json:"sideEffects"`
	Engines     struct {
		Npm  string `json:"npm"`
		Node string `json:"node"`
	} `json:"engines"`
	SimpleGitHooks  map[string]string   `json:"simple-git-hooks"`
	Main            string              `json:"main"`
	Author          string              `json:"author"`
	License         string              `json:"license"`
	Dependencies    map[string]string   `json:"dependencies"`
	Scripts         map[string]string   `json:"scripts"`
	DevDependencies map[string]string   `json:"devDependencies"`
	LintStaged      map[string][]string `json:"lint-staged"`
}

type Npm map[string]string
