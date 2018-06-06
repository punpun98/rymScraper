package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

//RockScrapegets a genre and if the genre exists it gets the top ten albums
func genreScrape(genre string) {
	client := &http.Client{}

	res, err := http.NewRequest("GET", "https://rateyourmusic.com/customchart?page=1&chart_type=top&type=album&year=alltime&genre_include=1&genres="+genre+"&include_child_genres=t&include=both&limit=none&countries=", nil)
	if err != nil {
		log.Fatal(err)
	}
	res.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0")

	req, err := client.Do(res)
	if req.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", req.StatusCode, req.Status)
	}
	defer req.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Top 40 albums in " + genre)
	doc.Find(".chart_detail").Each(func(i int, s *goquery.Selection) {
		band := s.Find(".chart_detail_line1 a").Text()
		title := s.Find(".chart_detail_line2 a").Text()
		fmt.Printf("Album %d: %s - %s\n", i+1, band, title)
	})
}
func main() {
	genreScrape("rock")
}
