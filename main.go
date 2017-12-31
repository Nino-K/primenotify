package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var conf = flag.String("config", "", "Path to a configuration file")

type Config struct {
	ExpectedRate float32  `json:"expected_rate"`
	URL          string   `json:"url"`
	From         string   `json:"from"`
	Recipients   []string `json:"recipients"`
	SMTPAddr     string   `json:"smtp_addr"`
	SMTPPort     int32    `json:"smtp_port"`
	SMTPAuth     string   `json:"smtp_auth"`
}

func main() {
	flag.Parse()
	if *conf == "" {
		log.Fatal("path to a configuration file must be privided")
	}
	config := new(Config)
	configFile, err := os.Open(*conf)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(config)
	c := contents(config.URL)
	currentRate := findPrimeRate(c)
	if currentRate != config.ExpectedRate {
		sendEmail(config, currentRate)
	}
}

func contents(url string) io.Reader {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Body
}

func findPrimeRate(body io.Reader) float32 {
	rateLine := extractRateLine(body)
	// find only floating numbers
	re := regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
	rates := re.FindAllString(rateLine, -1)

	// ignore the <h2> tag, should do this in regex above
	return convertRates(rates)[1]
}

func extractRateLine(body io.Reader) string {
	s := bufio.NewScanner(body)
	for s.Scan() {
		if strings.Contains(s.Text(), "td-copy-black td-margin-top-small td-margin-bottom-small td-copy-align-centre") {
			return s.Text()
		}
	}
	return ""
}

func convertRates(rates []string) []float32 {
	var fRates []float32
	for _, f := range rates {
		conv, err := strconv.ParseFloat(f, 32)
		if err == nil {
			fRates = append(fRates, float32(conv))
		}
	}
	return fRates
}

func sendEmail(config *Config, rate float32) {
	auth := smtp.PlainAuth("", config.From, config.SMTPAuth, config.SMTPAddr)
	msg := "From: " + config.From + "\n" +
		"To: " + strings.Join(config.Recipients, ",") + "\n" +
		"Subject: IMPORTANT PRIME RATE CHANGES\n\n" +
		"Prime rate today has gone up to" + fmt.Sprintf("%6.2f\n", rate)
	fmt.Println(msg)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.SMTPAddr, config.SMTPPort), auth, config.From, config.Recipients, []byte(msg))
	if err != nil {
		log.Fatal(err)
	}
}
