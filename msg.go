package main

type WebResponse struct {
}

type WebChunksResponse struct {
}

type HigherResponse struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type VideoResponse struct {
	Higher HigherResponse `json:"higher"`
}

type HighResponse struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type AudioResponse struct {
	High HighResponse `json:"high"`
}

type HTML5Response struct {
	Video VideoResponse `json:"video"`
	Audio AudioResponse `json:"audio"`
}

type IphoneResponse struct {
}

type MobileResponse struct {
	Video string `json:"video"`
}

type ShareResponse struct {
	Default string `json:"default"`
}

type FileVersionsResponse struct {
	Web       WebResponse       `json:"web"`
	WebChunks WebChunksResponse `json:"web_chunks"`
	HTML5     HTML5Response     `json:"html5"`
	Iphone    IphoneResponse    `json:"iphone"`
	Mobile    MobileResponse    `json:"mobile"`
	Share     ShareResponse     `json:"share"`
}

type AudioVersionsResponse struct {
}

type ImageVersionsResponse struct {
}

type FirstFrameVersionsResponse struct {
}

type DimensionsResponse struct {
}

type ChannelResponse struct {
}

type TagResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Value string `json:"value"`
}
type MediaBlocksResponse struct {
}

type CoubResponse struct {
	ID             int     `json:"Id"`
	Type           string  `json:"type"`
	Permalink      string  `json:"permalink"`
	Title          string  `json:"title"`
	VisibilityType string  `json:"visibility_type"`
	ChannelID      int     `json:"channel_id"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	IsDone         bool    `json:"is_done"`
	Duration       float64 `json:"duration"`
	ViewsCount     int     `json:"views_count"`
	Cotd           bool    `json:"cotd"`
	//CotdAt             int                        `json:"cotd_at"`
	Recoub             bool                       `json:"recoub"`
	Like               bool                       `json:"like"`
	RecoubsCount       int                        `json:"recoubs_count"`
	LikesCount         int                        `json:"likes_count"`
	RecoubTo           int                        `json:"recoub_to"`
	Flag               bool                       `json:"flag"`
	OriginalSound      bool                       `json:"original_sound"`
	HasSound           bool                       `json:"has_sound"`
	FileVersions       FileVersionsResponse       `json:"file_versions"`
	AudioVersions      AudioVersionsResponse      `json:"audio_versions"`
	ImageVersions      ImageVersionsResponse      `json:"image_versions"`
	FirstFrameVersions FirstFrameVersionsResponse `json:"first_frame_versions"`
	Dimensions         DimensionsResponse         `json:"dimensions"`
	AgeRestricted      bool                       `json:"age_restricted"`
	AllowReuse         bool                       `json:"allow_reuse"`
	Banned             bool                       `json:"banned"`
	//ExternalDownload     ExternalDownloadResponse   `json:"external_download"`
	Channel     ChannelResponse `json:"channel"`
	PercentDone int             `json:"percent_done"`
	Tags        []TagResponse   `json:"tags"`
	//RawVideoID           int                 `json:"raw_video_id"`
	MediaBlocks          MediaBlocksResponse `json:"media_blocks"`
	RawVideoThumbnailUrl string              `json:"raw_video_thumbnail_url"`
	RawVideoTitle        string              `json:"raw_video_title"`
	Meta                 string
}

type TimelineResponse struct {
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
	PerPage    int            `json:"per_page"`
	Coubs      []CoubResponse `json:"coubs"`
}
