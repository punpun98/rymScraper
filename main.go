package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	ms    = map[string]func(string, string, *discordgo.MessageCreate, *discordgo.Session){
		"!genre":  topGenreScrape,
		"!artist": artistScrape,
	}
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error bot machine broken", err)
		return

	}
	defer dg.Close()

	dg.AddHandler(musicMessage)

	if err = dg.Open(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func musicMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.Bot {
		return
	}
	if strings.HasPrefix(m.Content, "!") {
		var message = strings.Fields(m.Content)
		mes := message[1]
		query := message[1]
		fmt.Println(len(message))
		if len(message) > 2 {
			if query == "genre" {
				for i := 1; i <= len(message)-1; i++ {
					query += " " + message[i]
					mes += "+" + message[i]
				}
			} else if query == "artist" {
				for i := 1; i <= len(message)-1; i++ {
					query += " " + message[i]
					mes += "-" + message[i]
				}
				fmt.Println(query)
			}
		}
		if i, ok := ms[message[0]]; ok {
			i(query, mes, m, s)
		}

	}

}

//RockScrapegets a genre and if the genre exists it gets the top ten albums
func topGenreScrape(genre string, message string, m *discordgo.MessageCreate, s *discordgo.Session) {
	client := &http.Client{}
	var mes string
	res, err := http.NewRequest("GET", "https://rateyourmusic.com/customchart?page=1&chart_type=top&type=album&year=alltime&genre_include=1&genres="+message+"&include_child_genres=t&include=both&limit=10&countries=", nil)
	if err != nil {
		log.Fatal(err)
	}
	res.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/45.0")

	req, err := client.Do(res)
	if req.StatusCode != 200 {
		s.ChannelMessageSend(m.ChannelID, "RYM has probably banned me for a bit. Please hold on ")
		log.Fatalf("Status code error: %d %s", req.StatusCode, req.Status)
	}
	defer req.Body.Close()

	doc, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	mes += "Top ten albums in " + genre + "\n\n"
	fmt.Println(doc.Find(".error").Text())
	if (strings.Contains(doc.Find(".error").Text(), "The following genres were not found and therefore ignored: ")) == false {
		doc.Find(".chart_detail").EachWithBreak(func(i int, s *goquery.Selection) bool {
			band := s.Find(".chart_detail_line1 a").Text()
			title := s.Find(".chart_detail_line2 a").Text()
			mes += fmt.Sprintf("%d: %s - %s\n", i+1, band, title)
			if i >= 9 {
				return false
			}
			return true
		})
	} else {
		mes = "Could Not Find That Genre. ¯\\_(ツ)_/¯ "
	}
	s.ChannelMessageSend(m.ChannelID, mes)

}

func artistScrape(artist string, message string, m *discordgo.MessageCreate, s *discordgo.Session) {
	client := &http.Client{}
	var mes string
	fmt.Println("https://rateyourmusic.com/artist/" + message)
	res, err := http.NewRequest("GET", "https://rateyourmusic.com/artist/"+message, nil)
	if err != nil {
		log.Fatal(err)
	}
	res.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/45.0")

	req, err := client.Do(res)
	if req.StatusCode == 404 {
		s.ChannelMessageSend(m.ChannelID, "Artist doesn't exist. Soz")
		log.Fatalf("Status code error: %d %s", req.StatusCode, req.Status)
	}
	if req.StatusCode != 200 {
		s.ChannelMessageSend(m.ChannelID, "RYM has probably banned me for a bit. Please hold on")
		log.Fatalf("Status code error: %d %s", req.StatusCode, req.Status)
	}
	defer req.Body.Close()

	doc, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	mes += "Top albums by " + artist + "\n\n"
	fmt.Println(doc.Find(".error").Text())
	if (strings.Contains(doc.Find(".error").Text(), "The following genres were not found and therefore ignored: ")) == false {
		doc.Find(".disco_info").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".album").Text()
			year := s.Find(".disco_year_ymd").Text()
			mes += fmt.Sprintf("%d: %s - %s\n", i+1, title, year)
		})
	} else {
		mes = "Could Not Find That Genre. ¯\\_(ツ)_/¯ "
	}
	s.ChannelMessageSend(m.ChannelID, mes)
}
