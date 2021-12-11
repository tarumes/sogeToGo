package types

type GithubApiRelease struct {
	Url       string `json:"url"`
	AssetsUrl string `json:"assets_url"`
	UploadUrl string `json:"upload_url"`
	HtmlUrl   string `json:"html_url"`
	Id        int64  `json:"id"`
	Author    struct {
		Login string `json:"login"`
		Id    int64  `json:"id"`
		//...
	} `json:"author"`

	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
	TarballUrl string `json:"tarball_url"`
	ZipballUrl string `json:"zipball_url"`
	Body       string `json:"body"`
}
