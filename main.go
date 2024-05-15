package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
)

// A Response struct to map the Entire Response
type Response struct {
	Name    string    `json:"name"`
	Pokemon []Pokemon `json:"pokemon_entries"`
}

// A Pokemon Struct to map every pokemon to.
type Pokemon struct {
	EntryNo int            `json:"entry_number"`
	Species PokemonSpecies `json:"pokemon_species"`
}

// A struct to map our Pokemon's Species which includes it's name
type PokemonSpecies struct {
	Name string `json:"name"`
}

var firebaseClient *firestore.Client = new(firestore.Client)
var apiKey string
var networkId string

func pokemon() {
	response, err := http.Get("http://pokeapi.co/api/v2/pokedex/kanto/")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject Response
	json.Unmarshal(responseData, &responseObject)

	fmt.Println(responseObject.Name)
	fmt.Println(len(responseObject.Pokemon))

	for i := 0; i < len(responseObject.Pokemon); i++ {
		fmt.Println(responseObject.Pokemon[i].Species.Name)
	}

}

func printMerakiNetworkName(w http.ResponseWriter, r *http.Request) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.meraki.com/api/v1/networks/%s", networkId), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Cisco-Meraki-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var network map[string]interface{}
	json.Unmarshal(body, &network)

	//generate a byte array of 5
	randomBytes := make([]byte, 5)
	_, err = rand.Read(randomBytes)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Fprintf(w, "Bytes: %s", randomBytes)
	cellId := base32.StdEncoding.EncodeToString(randomBytes)
	fmt.Printf("Network Name: %s \n", network["name"])
	fmt.Fprintf(w, "Network Name: %s, cell ID %s \n", network["name"], cellId)
}

func helloworld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func init() {

}

func main() {
	err := godotenv.Load() // 👈 load .env file
	if err != nil {
		log.Fatal(err)
	}

	tgBotToken := os.Getenv("TG_BOT_TOKEN")

	apiKey = os.Getenv("MERAKI_API_KEY")
	networkId = os.Getenv("MERAKI_NETWORK_ID")

	// fmt.Println(tgBotToken)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fmt.Println("Preparing database")

	sa := option.WithCredentialsFile("./merakle-dev-firebase-adminsdk-vlqb4-65dba08b77.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	firebaseClient, err := app.Firestore(ctx)
	if err != nil {
		fmt.Printf("Error connecting to firebase:\n")
		log.Fatalln(err)
	}
	defer firebaseClient.Close()
	fmt.Printf("Client: %s\n", firebaseClient)
	fmt.Printf("Connected to firebase:\n")

	iter := firebaseClient.Collection("cells").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Println(doc.Data())
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(tgBotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)

	// http.HandleFunc("/", printMerakiNetworkName)
	// http.ListenAndServe(":8080", nil)
}

func fetchCellsByTelegramUsername(client *firestore.Client, ctx context.Context, username string) {
	fmt.Printf("Fetching cells for user %s\n", username)
	// ctx := context.Background()
	fmt.Printf("Client: %s\n", firebaseClient)

	var collectionsIter = firebaseClient.Collections(ctx)
	for {
		col, err := collectionsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Println(col.ID)
	}
	// fmt.Printf("Test ID: %s\n", testId)

	dsnap, err := firebaseClient.Collection("cells").Doc("clfuRxavPnajb2aewP6N").Get(ctx)
	if err != nil {
		log.Printf("Failed to get document: %v", err)
		return
	}

	var cells []string
	dsnap.DataTo(&cells)

	fmt.Printf("Cells for user %s: %v\n", username, cells)

}

func initFirebase() {

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// b.SendMessage(ctx, &bot.SendMessageParams{
	// 	ChatID: update.Message.Chat.ID,
	// 	Text:   update.Message.Text,
	// })
	//user, _ := b.GetMe(context.Background())

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Button 1", CallbackData: "button_1"},
				{Text: "Button 2", CallbackData: "button_2"},
			}, {
				{Text: "Button 3", CallbackData: "button_3"},
			},
		},
	}

	incomingUsername := update.Message.From.Username
	fetchCellsByTelegramUsername(firebaseClient, ctx, incomingUsername)
	if incomingUsername == "KonstantinIlinov" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Hello, KonstantinIlinov",
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Click by button",
			ReplyMarkup: kb,
		})
	}
	fmt.Printf("Received message from %s: %s\n", update.Message.From.Username, update.Message.Text)
}
