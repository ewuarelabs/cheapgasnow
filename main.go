package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const URL = "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey="

var lastGas int
var counter = 0

func GetGas() string {

	authURL := URL + os.Getenv("API_KEY")
	resp, err := http.Get(authURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	res := struct {
		Status  string
		Message string
		Result  struct {
			LastBlock        string
			SafeGasPrice     string
			ProposeGasPrice  string
			FastGasPrice     string
			SuggestedBaseFee string
			GasUsedRatio     string
		}
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatal("ooopsss! an error occurred, please try again")
	}
	return res.Result.ProposeGasPrice
}

func BuildTweet(price string) string {

	tweet1 := "gas is low, currently at " + price + " gwei! knock yourself out!"
	tweet2 := "gas is " + price + "gwei!"
	tweet3 := "it's pretty lowwww right now at " + price + " gwei"
	tweet4 := "gas prices are pretty chill: " + price + " gwei"

	tweets := make([]string, 0)
	tweets = append(tweets,
		tweet1,
		tweet2,
		tweet3,
		tweet4)

	//select at random
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	tweet := fmt.Sprint("Gm, ", tweets[rand.Intn(len(tweets))])

	return tweet
}

func Job() {
	c := cron.New()
	c.AddFunc("@every 30s", func() {
		fmt.Println("starting Job!")
		gas := GetGas()


		intGas, err := strconv.Atoi(gas)
		if err != nil {
			log.Fatal("error converting the returned gas price from a string to an integer")
		}

		if intGas < 80 && intGas > 30 && counter == 0 {
			lastGas = intGas
			newTweet := BuildTweet(gas)
			fmt.Printf("Gas is currently %s gwei", gas)
			SendTweet(newTweet)
			counter++
			time.Sleep(5 * time.Second)

		} else if counter > 0 && intGas < 80 && intGas > 30 {
			percentage := ((lastGas - intGas) / lastGas) * 100
			if math.Abs(float64(percentage)) > 7.00 {
				deviatedTweet := fmt.Sprintf("gas prices have deviated significantly from the last price, the current gas price is %v gwei", intGas)
				SendTweet(deviatedTweet)
			}
			counter = 0
			fmt.Printf("Gas is currently %s gwei\n", gas)
		}
		fmt.Printf("Gas is currently %s gwei\n", gas)
	})
	c.Start()
}

func SendTweet(gastweet string) {

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	log.Println(&httpClient)
	// twitter client
	client := twitter.NewClient(httpClient)
	log.Println(client)

	tweet, resp, err := client.Statuses.Update(gastweet, nil)
	fmt.Println(*resp)
	fmt.Println(*tweet)
	if err != nil {
		log.Println(err)
	}

}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	r := mux.NewRouter().StrictSlash(true)
	Job()

	gasf := GetGas()
	fmt.Println("Gas is " + gasf)

	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(8000)
	}
	fmt.Printf("Listening and serving on port %s.....\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
