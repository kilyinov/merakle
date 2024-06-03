package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"google.golang.org/api/iterator"
)

func StartBot(ctx context.Context, tgBotToken string, firebaseClient *firestore.Client) {
	print("StartBot")
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(tgBotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		fmt.Printf("Callback query: %s\n", update.CallbackQuery.Data)
		callbackQuery := update.CallbackQuery
		callbackQueryCommand := update.CallbackQuery.Data
		switch callbackQueryCommand {
		case "bCreateCell":
			newCellNameKB := &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{Text: "Yes", CallbackData: "bCreateCell"},
					},
				},
			}

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      callbackQuery.Message.Message.Chat.ID,
				Text:        "Please enter the name of your new cell",
				ReplyMarkup: newCellNameKB},
			)
		}
	} else {
		createCellKB := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Yes", CallbackData: "bCreateCell"},
					{Text: "No", CallbackData: "bCancel"},
				},
			},
		}

		incomingUsername := update.Message.From.Username
		userCells, found := fetchCellsByTelegramUsername(ctx, incomingUsername)
		if found {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Welcome back! Here are your cells: " + fmt.Sprintf("%#v", userCells),
				ReplyMarkup: createCellKB,
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Looks like you don't have any cells yet. Would you like to create one?",
				ReplyMarkup: createCellKB,
			})
		}
		fmt.Printf("Received message from %s: %s\n", update.Message.From.Username, update.Message.Text)
	}
}

func fetchCellsByTelegramUsername(ctx context.Context, username string) ([]Cell, bool) {
	fmt.Printf("Fetching cells for user %s\n", username)
	var cellCollection = firebaseClient.Collection("cells")

	query := cellCollection.Where("admins", "array-contains", username).Documents(ctx)
	var cells []Cell
	for {
		cell, err := query.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		jsonbody, err := json.Marshal(cell.Data())
		if err != nil {
			log.Fatalf("Failed to parse json from database: %v", err)
		}
		cellObject := Cell{}
		if err := json.Unmarshal(jsonbody, &cellObject); err != nil {
			// do error check
			fmt.Println(err)
		}
		cellObject.ID = cell.Ref.ID

		fmt.Printf("%#v\n", cellObject)
		cells = append(cells, cellObject)
	}
	return cells, len(cells) > 0
}
