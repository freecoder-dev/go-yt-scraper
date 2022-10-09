package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/gocolly/colly"
	"gopkg.in/yaml.v3"
)

const database = "result.csv"

var kUserAgent = []string{
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:24.0) Gecko/20100101 Firefox/24.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/91.0.4472.114 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:77.0) Gecko/20100101 Firefox/77.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:77.0) Gecko/20100101 Firefox/77.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/78.0.3904.70 Safari/537.36",
	"Mozilla/5.0 (X11; Linux i586; rv:31.0) Gecko/20100101 Firefox/31.0"}

type ConfigData struct {
	SearchTerms      string `yaml:"search_terms"`
	GeolocationTerms string `yaml:"geo_locaction_terms"`
	MaxPages         int    `yaml:"max_pages"`
	RequestTimes     int    `yaml:"request_times"`
}

func main() {
	fmt.Println("YT scraper..")
	fmt.Println()
	configData := &ConfigData{}

	yamlFile, err := os.ReadFile("config.yaml")
	if err == nil {
		_ = yaml.Unmarshal(yamlFile, configData)
	} else {
		return
	}

	var (
		kSearchTerms      = url.QueryEscape(configData.SearchTerms)
		kGeolocationTerms = url.QueryEscape(configData.GeolocationTerms)
		kUrlFormat        = "https://www.yellowpages.com/search?search_terms=%s&geo_location_terms=%s&page=%d"
	)

	file, err := os.Create(database)
	if err != nil {
		log.Fatalf("erreur of creating file %s", err)
		return
	}
	defer file.Close()

	dbWriter := csv.NewWriter(file)
	dbWriter.Comma = ';'
	dbWriter.Write([]string{
		"Business Name",
		"Website",
		"Telephone",
		"Address",
	})
	defer dbWriter.Flush()

	fmt.Printf("# + Query: %s | %s    #\n", kSearchTerms, kGeolocationTerms)

	for idx := 1; idx < configData.MaxPages+1; idx++ {
		userAgentId := rand.Intn(cap(kUserAgent))

		c := colly.NewCollector(colly.UserAgent(kUserAgent[userAgentId]))
		c.SetRequestTimeout(120 * time.Second)
		c.OnHTML(".info", func(e *colly.HTMLElement) {
			if e.ChildText("div.info-section.info-primary > h2 > a > span") != "" {
				dbWriter.Write([]string{
					e.ChildText("div.info-section.info-primary > h2 > a > span"),
					e.ChildAttr("div.info-section.info-primary > div.links > a.track-visit-website", "href"),
					e.ChildText("div.info-section.info-secondary > div.phones.phone.primary"),
					e.ChildText("div.info-section.info-secondary > div.adr"),
				})
			}
		})

		url := fmt.Sprintf(kUrlFormat, kSearchTerms, kGeolocationTerms, idx)
		fmt.Printf("\r# +Scraping page number %.2d  ---> DONE        #", idx)
		c.Visit(url)
		time.Sleep(time.Duration(configData.RequestTimes) * time.Second)
	}
}
