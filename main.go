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
		"!genre": topGenreScrape,
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
		genre := message[1]
		fmt.Println(len(message))
		if len(message) > 1 {
			for i := 2; i <= len(message)-1; i++ {
				genre += " " + message[i]
				mes += "+" + message[i]
			}
		}
		if i, ok := ms[message[0]]; ok {
			i(genre, mes, m, s)
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
	res.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0")

	req, err := client.Do(res)
	if req.StatusCode != 200 {
		s.ChannelMessageSend(m.ChannelID, "Ay yo pick some real music")
		log.Fatalf("Status code error: %d %s", req.StatusCode, req.Status)
	}
	defer req.Body.Close()

	doc, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	mes += "Top ten albums in " + genre + "\n\n"
	fmt.Println(doc.Find(".error").Text())
	if doc.Find(".error").Text() != "The following genres were not found and therefore ignored: "+genre {
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
