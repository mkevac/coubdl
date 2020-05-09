package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/browser"
	"github.com/schollz/progressbar"
	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

const configFileName = "settings.json"

type ConfigStruct struct {
	ApplicationID string `json:"application_id"`
	Secret        string `json:"secret"`
	Token         string `json:"token"`
}

var (
	config         ConfigStruct
	destinationDir string
	limit          int
)

func authenticate(applicationID string, secret string) (string, error) {
	var (
		codeChan        = make(chan string)
		coubOauthConfig *oauth2.Config
	)

	coubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     applicationID,
		ClientSecret: secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://coub.com/oauth/authorize",
			TokenURL: "http://coub.com/oauth/token",
		},
	}

	url := coubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	if err := browser.OpenURL(url); err != nil {
		fmt.Printf("Error while opening your browser. Please go to URL %s manually.", url)
		return "", err
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		codeChan <- r.FormValue("code")
		fmt.Fprintf(w, "You can now close the browser.")
	})
	server := http.Server{
		Addr: ":8080",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("error while listening on ':8080': %s", err)
			}
			return
		}
	}()

	code := <-codeChan

	// we don't need http server any more
	if err := server.Close(); err != nil {
		return "", fmt.Errorf("error while closing server: %s", err)
	}

	token, err := coubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Fatal(err)
	}

	return token.AccessToken, nil
}

func getNumberOfCoubs(token string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("http://coub.com/api/v2/timeline/likes?access_token=%s&per_page=1&page=1", token))
	if err != nil {
		return 0, fmt.Errorf("error while doing http req: %s", err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error while reading http response: %s", err)
	}

	timeline := TimelineResponse{}

	if err := json.Unmarshal(bytes, &timeline); err != nil {
		return 0, fmt.Errorf("error while unmarshalling response json: %s", err)
	}

	return timeline.TotalPages, nil
}

func getCoubsFromSite(token string, limit int) ([]CoubResponse, error) {
	fmt.Println("Downloading coub list from site...")

	totalCoubs, err := getNumberOfCoubs(token)
	if err != nil {
		return nil, fmt.Errorf("error while getting number of coubs: %w", err)
	}

	if limit == 0 {
		limit = totalCoubs
	}

	var coubs []CoubResponse

	bar := progressbar.New(limit)

	for page := 1; len(coubs) < totalCoubs && len(coubs) < limit; page++ {

		resp, err := http.Get(fmt.Sprintf("http://coub.com/api/v2/timeline/likes?access_token=%s&per_page=50&page=%d", token, page))
		if err != nil {
			return nil, fmt.Errorf("error while doing http req: %s", err)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("error while reading http response: %s", err)
		}

		resp.Body.Close()

		timeline := TimelineResponse{}

		if err := json.Unmarshal(bytes, &timeline); err != nil {
			return nil, fmt.Errorf("error while unmarshalling response json: %s", err)
		}

		if len(timeline.Coubs) == 0 {
			break
		}

		added := 0
		for _, coub := range timeline.Coubs {
			if len(coubs) < limit {
				coubs = append(coubs, coub)
				added++
			}
		}
		bar.Add(added)
	}

	fmt.Println()

	return coubs, nil
}

func downloadFile(url string, size int, name string, destinationDir string) error {
	finalPath := filepath.Join(destinationDir, name)

	stat, err := os.Stat(finalPath)
	if err == nil {
		if size != 0 && stat.Size() == int64(size) { // already downloaded
			return nil
		} else { // already downloaded
			return nil
		}
	}

	out, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("error while creating file '%s': %w", finalPath, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error while getting url '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error while getting url '%s': bad status: %s", url, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error while downloading file from url '%s' to '%s': %w", url, finalPath, err)
	}

	return nil
}

func saveIndexFile(coubs []CoubResponse, destinationDir string) error {
	if err := os.MkdirAll(destinationDir, 0777); err != nil {
		return fmt.Errorf("error while creating destination dir '%s': %w", destinationDir, err)
	}

	out, err := os.Create(filepath.Join(destinationDir, "index.html"))
	if err != nil {
		return fmt.Errorf("error while creating index.html file: %w", err)
	}
	defer out.Close()

	fmt.Fprintf(out, "<!doctype html>\n")
	fmt.Fprintf(out, "<html lang=\"en\">\n")
	fmt.Fprintf(out, "<head>\n")
	fmt.Fprintf(out, "<meta charset=\"utf-8\">")
	fmt.Fprintf(out, "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">")
	fmt.Fprintf(out, "<link rel=\"stylesheet\" href=\"https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css\" integrity=\"sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh\" crossorigin=\"anonymous\">")
	fmt.Fprintf(out, "<title>My Favorite Coubs</title>\n")
	fmt.Fprintf(out, "</head>\n")
	fmt.Fprintf(out, "<body>\n")

	fmt.Fprintf(out, "<div class=\"container\">\n")
	fmt.Fprintf(out, "<div class=\"row\">\n")
	fmt.Fprintf(out, "<h1>My Favorite Coubs</h1>")
	fmt.Fprintf(out, "</div>\n")

	fmt.Fprintf(out, "<div class=\"row\">\n")
	fmt.Fprintf(out, "<div class=\"col\">\n")

	fmt.Fprintf(out, "<table class=\"table table-striped table-hover table-sm\">\n")

	fmt.Fprintf(out, "<thead>\n")
	fmt.Fprintf(out, "<th scope=\"col\">#</th>\n")
	fmt.Fprintf(out, "<th scope=\"col\">Name</th>\n")
	fmt.Fprintf(out, "<th scope=\"col\">Tags</th>\n")
	fmt.Fprintf(out, "<th scope=\"col\">Video</th>\n")
	fmt.Fprintf(out, "</thead>\n")

	fmt.Fprintf(out, "<tbody>\n")

	for no, coub := range coubs {
		fmt.Fprintf(out, "<tr>\n")
		fmt.Fprintf(out, "<th scope=\"row\">%d</td>\n", no)
		fmt.Fprintf(out, "<td><a href=\"https://coub.com/view/%s\">%s</a></td>\n", coub.Permalink, coub.Title)
		fmt.Fprintf(out, "<td>\n")
		for i, tag := range coub.Tags {
			if i != 0 {
				fmt.Fprintf(out, " ")
			}
			fmt.Fprintf(out, "<div class=\"badge badge-primary\">%s</div>", tag.Title)
		}
		fmt.Fprintf(out, "</td>\n")
		fmt.Fprintf(out, "<td>\n")
		fmt.Fprintf(out, "<video width=\"320\" height=\"240\" controls loop><source src=\"%s\" type=\"video/mp4\"></video>", "./"+getFileName(coub, ""))
		fmt.Fprintf(out, "</td>\n")
		fmt.Fprintf(out, "</tr>\n")
	}

	fmt.Fprintf(out, "<tbody>\n")

	fmt.Fprintf(out, "</table>\n")

	fmt.Fprintf(out, "</div>\n")
	fmt.Fprintf(out, "</div>\n")
	fmt.Fprintf(out, "</div>\n")

	fmt.Fprintf(out, "</body>\n")
	fmt.Fprintf(out, "</html>\n")

	return nil
}

func sanitizeTitle(title string) string {
	return strings.ReplaceAll(title, "/", "-")
}

func getFileName(coub CoubResponse, suffix string) string {
	url := coub.FileVersions.Share.Default
	extension := filepath.Ext(url)
	name := sanitizeTitle(coub.Title) + " (" + coub.Permalink + ")" + suffix + extension
	return name
}

func downloadCoubs(coubs []CoubResponse, destinationDir string) error {
	if err := os.MkdirAll(destinationDir, 0777); err != nil {
		return fmt.Errorf("error while creating destination dir '%s': %w", destinationDir, err)
	}

	fmt.Printf("Downloading %d coubs to '%s'...\n", len(coubs), destinationDir)

	bar := progressbar.New(len(coubs) * 3) // looped + video + audio

	for _, coub := range coubs {

		// looped med
		if coub.FileVersions.Share.Default != "" {
			name := getFileName(coub, "")
			if err := downloadFile(coub.FileVersions.Share.Default, 0, name, destinationDir); err != nil {
				return err
			}
		} else {
			//log.Printf("no looped url for coub '%s'", coub.Title)
		}
		bar.Add(1)

		// good video
		if coub.FileVersions.HTML5.Video.Higher.URL != "" {
			name := getFileName(coub, "_video")
			if err := downloadFile(coub.FileVersions.HTML5.Video.Higher.URL, coub.FileVersions.HTML5.Video.Higher.Size, name, destinationDir); err != nil {
				return err
			}
		} else {
			//log.Printf("no video url for coub '%s'", coub.Title)
		}
		bar.Add(1)

		// good sound
		if coub.FileVersions.HTML5.Audio.High.URL != "" {
			name := getFileName(coub, "_audio")
			if err := downloadFile(coub.FileVersions.HTML5.Audio.High.URL, coub.FileVersions.HTML5.Audio.High.Size, name, destinationDir); err != nil {
				return err
			}
		} else {
			//log.Printf("no audio url for coub '%s'", coub.Title)
		}
		bar.Add(1)

	}

	fmt.Println()

	return nil
}

func populateConfig() error {
	configDirs := configdir.New("mkevac", "coubdl")
	folder := configDirs.QueryFolderContainsFile(configFileName)
	if folder != nil {
		data, err := folder.ReadFile(configFileName)
		if err == nil {
			json.Unmarshal(data, &config)
		}
	}

	if config.Token != "" { // having just token is enough
		return nil
	}

	// ok, token is empty, but maybe we have application_id and secret?

	if config.ApplicationID != "" && config.Secret != "" {
		token, err := authenticate(config.ApplicationID, config.Secret)
		if err != nil {
			return fmt.Errorf("error while authenticating with given application_id and secret: %w", err)
		}
		config.Token = token

		data, err := json.Marshal(&config)
		if err != nil {
			log.Fatalf("error while marshalling config to json: %s", err)
		}

		folders := configDirs.QueryFolders(configdir.Global)
		folders[0].WriteFile(configFileName, data)
	}

	// ask for application_id and secret

	var applicationID string
	var secret string

	for applicationID == "" {
		fmt.Printf("Please enter Application ID: ")
		fmt.Scanln(&applicationID)
		if applicationID == "" {
			fmt.Println("Application ID is empty.")
		}
	}

	for secret == "" {
		fmt.Printf("Please enter Secret: ")
		fmt.Scanln(&secret)
		if secret == "" {
			fmt.Println("Secret is empty.")
		}
	}

	fmt.Println("You have entered")
	fmt.Printf("Application ID: %s\n", applicationID)
	fmt.Printf("Secret: %s\n", secret)

	config.ApplicationID = applicationID
	config.Secret = secret

	token, err := authenticate(config.ApplicationID, config.Secret)
	if err != nil {
		return fmt.Errorf("error while authenticating with given application_id and secret: %w", err)
	}
	config.Token = token

	data, err := json.Marshal(&config)
	if err != nil {
		log.Fatalf("error while marshalling config to json: %s", err)
	}

	folders := configDirs.QueryFolders(configdir.Global)
	folders[0].WriteFile(configFileName, data)

	return nil
}

func main() {
	flag.StringVar(&destinationDir, "dir", ".", "destination directory for downloading")
	flag.IntVar(&limit, "limit", 0, "limit number of coubs to download by this number (0 - no limit)")
	flag.Parse()

	if err := populateConfig(); err != nil {
		log.Fatalf("error while populating config: %s", err)
	}

	coubs, err := getCoubsFromSite(config.Token, limit)
	if err != nil {
		log.Fatalf("error while getting coubs from site: %s", err)
	}

	if err := downloadCoubs(coubs, destinationDir); err != nil {
		log.Fatalf("error while dowloading coubs: %s", err)
	}

	if err := saveIndexFile(coubs, destinationDir); err != nil {
		log.Fatalf("error while saving index file: %s", err)
	}

	browser.OpenURL(fmt.Sprintf("file:///%s/index.html", destinationDir))
}
