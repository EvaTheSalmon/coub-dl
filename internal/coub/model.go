package coub

type CoubResponse struct {
	Page       int    `json:"page"`
	TotalPages int    `json:"total_pages"`
	Coubs      []Coub `json:"coubs"`
}

type Coub struct {
	Permalink    string       `json:"permalink"`
	Title        string       `json:"title"`
	UpdatedAt    string       `json:"updated_at"`
	Channel      Channel      `json:"channel"`
	Tags         []Tag        `json:"tags"`
	FileVersions FileVersions `json:"file_versions"`
}

type Channel struct {
	Title     string `json:"title"`
	Permalink string `json:"permalink"`
}

type Tag struct {
	Title string `json:"title"`
}

type FileVersions struct {
	HTML5 HTML5 `json:"html5"`
}

type HTML5 struct {
	Video MediaTiers `json:"video"`
	Audio MediaTiers `json:"audio"`
}

type MediaTiers struct {
	Higher MediaVariant `json:"higher"`
	High   MediaVariant `json:"high"`
	Med    MediaVariant `json:"med"`
}

type MediaVariant struct {
	URL string `json:"url"`
}
