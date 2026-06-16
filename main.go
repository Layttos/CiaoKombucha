package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bot.ciaokombucha.tv/Listener"
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		fmt.Println("An error occured while attemping to load the .env file (check if it doesn't exist). Just so you know, the program doesn't stop as the variables may be defined by the system.")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("Problem: The token wasn't found. Closing program...")
		return
	}

	fmt.Println("Connecting to local database...")
	Utils.DB, err = sql.Open("sqlite", "./ciaokombucha.db")
	if err != nil {
		log.Fatal(err)
	}
	defer Utils.DB.Close()

	messages_query := `CREATE TABLE IF NOT EXISTS messages(
		id TEXT PRIMARY KEY,
		channel_id TEXT,
		content TEXT,
		author_id TEXT
	);`

	deleted_msg_query := `CREATE TABLE IF NOT EXISTS deleted_messages(
		channel_id TEXT PRIMARY KEY,
		message TEXT,
		author_id TEXT
	);`



	Utils.DB.Query(messages_query)
	Utils.DB.Query(deleted_msg_query)




	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("An error occured while attemping to initiate the Discord bot session:", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMembers

	dg.AddHandler(Listener.MessageUpdate)
	dg.AddHandler(Listener.MessageCreate)
	dg.AddHandler(Listener.MemberUpdate)
	dg.AddHandler(Listener.MemberUpdateTag)
	dg.AddHandler(Listener.MemberBanned)
	dg.AddHandler(Listener.MemberKicked)
	dg.AddHandler(Listener.MessageDelete)
	dg.AddHandler(Listener.MemberJoin)
	dg.AddHandler(Listener.MemberQuit)


	err = dg.Open()
	if err != nil {
		fmt.Println("An error occured while attemping to start the discord bot:", err)
	}


	fmt.Println("Ciao Kombucha, en ligne !")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()


}