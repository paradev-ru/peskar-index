package movie

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
)

type Movie struct {
	URL         *url.URL
	Directory   string
	Name        string
	Trailer     string
	Cover       string
	Keywords    string
	Description string
	Image       string
	ImageWidth  string
	ImageHeight string

	rawurl string
}

const (
	RetraingMax       = 100
	KinopoiskPlusHost = "plus.kinopoisk.ru"
	UserAgent         = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

var (
	KinopoiskRegexp   = regexp.MustCompile(`https\:\/\/(.*)\.kinopoisk\.ru\/film\/([0-9]+)\/`)
	NameRegexp        = regexp.MustCompile(`itemprop\=\"name\"\>([^\<]*)`)
	DescriptionRegexp = regexp.MustCompile(`itemprop\=\"description\"\>([^\<]*)`)
	TrailerRegexp     = regexp.MustCompile(`content\=\"([^\"]*)\"\ property\=\"og\:video\:url\"`)
	ImageRegexp       = regexp.MustCompile(`content\=\"([^\"]*)\"\ property\=\"og\:image\"`)
	ImageWidthRegexp  = regexp.MustCompile(`content\=\"([0-9]+)\"\ property\=\"og\:image\:width\"`)
	ImageHeightRegexp = regexp.MustCompile(`content\=\"([0-9]+)\"\ property\=\"og\:image\:height\"`)
	KeywordsRegexp    = regexp.MustCompile(`content\=\"([^\"]*)\"\ name\=\"keywords\"`)
	CoverRegexp       = regexp.MustCompile(`class="image__source" srcset="([^\ ]*)\ 1x`)
)

func New(rawurl string) (*Movie, error) {
	if rawurl == "" {
		return nil, errors.New("Empty link")
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if KinopoiskRegexp.FindString(rawurl) == "" {
		return nil, errors.New("Get non Kinopoisk link: " + rawurl)
	}
	u.Host = KinopoiskPlusHost
	return &Movie{
		URL:    u,
		rawurl: rawurl,
	}, nil
}

func (m *Movie) Parse() error {
	var retraing int
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL:    m.URL,
	}
	request.Header.Set("User-Agent", UserAgent)
TRY:
	retraing++
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		if retraing == RetraingMax {
			return err
		}
		logrus.Debugf("Retraing %d/%d...", retraing, RetraingMax)
		time.Sleep(5 * time.Second)
		goto TRY
	}
	if resp.StatusCode != 200 {
		return errors.New("Get info error non 200 reponse: " + resp.Status)
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	s := string(html)
	nameData := NameRegexp.FindStringSubmatch(s)
	if len(nameData) != 0 {
		m.Name = nameData[1]
	}
	trailerData := TrailerRegexp.FindStringSubmatch(s)
	if len(trailerData) != 0 {
		m.Trailer = trailerData[1]
	}
	descriptionData := DescriptionRegexp.FindStringSubmatch(s)
	if len(descriptionData) != 0 {
		m.Description = descriptionData[1]
	}
	imageData := ImageRegexp.FindStringSubmatch(s)
	if len(imageData) != 0 {
		m.Image = imageData[1]
	}
	imageWidthData := ImageWidthRegexp.FindStringSubmatch(s)
	if len(imageWidthData) != 0 {
		m.ImageWidth = imageWidthData[1]
	}
	imageHeightData := ImageHeightRegexp.FindStringSubmatch(s)
	if len(imageHeightData) != 0 {
		m.ImageHeight = imageHeightData[1]
	}
	keywordsData := KeywordsRegexp.FindStringSubmatch(s)
	if len(keywordsData) != 0 {
		m.Keywords = keywordsData[1]
	}
	coverData := CoverRegexp.FindStringSubmatch(s)
	if len(coverData) != 0 {
		m.Cover = "http:" + coverData[1]
	}
	return nil
}

func (m *Movie) Template(templateFile string) (string, error) {
	var doc bytes.Buffer
	t := template.New("movie_template")
	t = template.Must(template.ParseFiles(templateFile))
	err := t.Execute(&doc, m)
	if err != nil {
		return "", err
	}
	return doc.String(), nil
}
