package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot.ciaokombucha.tv/Command"
	"bot.ciaokombucha.tv/Listener"
	"bot.ciaokombucha.tv/Radio"
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"

	_ "modernc.org/sqlite"
)

func RegisterCommand(cmd Utils.Command) {
	Utils.Commands = append(Utils.Commands, cmd)
}

func LoadCommands(s *discordgo.Session) {
	for _, cmd := range Utils.Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, os.Getenv("GUILD_ID"), &discordgo.ApplicationCommand{
			Name:                     cmd.Name(),
			Description:              cmd.Description(),
			DefaultMemberPermissions: cmd.Permissions(),
			Options:                  cmd.Options(),
		})
		if err != nil {
			fmt.Println("An error occured while attemping to register the command", cmd.Name(), ":", err)
		}
		fmt.Println("Command " + cmd.Name() + " est enregistrée.")
	}
}

func ConnectToRadioChannel(s *discordgo.Session) {
	Radio.ConnectToRadioChannel(s)
}

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

	if os.Getenv("IS_DEV") == "true" {
		token = os.Getenv("DISCORD_TOKEN_DEV")
		fmt.Println("Running in development mode.")
	}

	fmt.Println("Connecting to local database...")
	Utils.DB, err = sql.Open("sqlite", "./ciaokombucha.db")
	if err != nil {
		log.Fatal(err)
	}
	Utils.DB.SetMaxOpenConns(1)
	Utils.DB.SetMaxIdleConns(1)
	if _, err := Utils.DB.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		log.Fatal(err)
	}
	if _, err := Utils.DB.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
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

	management_query := `CREATE TABLE IF NOT EXISTS management(
		roles_message TEXT
	);`

	experience_query := `CREATE TABLE IF NOT EXISTS levels(
		user_id TEXT PRIMARY KEY,
		experience INTEGER,
		level INTEGER
	);`

	if _, err := Utils.DB.Exec(messages_query); err != nil {
		log.Fatal(err)
	}
	if _, err := Utils.DB.Exec(deleted_msg_query); err != nil {
		log.Fatal(err)
	}
	if _, err := Utils.DB.Exec(management_query); err != nil {
		log.Fatal(err)
	}
	if _, err := Utils.DB.Exec(experience_query); err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("An error occured while attemping to initiate the Discord bot session:", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
	dg.State.TrackVoice = true

	// Member Handler
	dg.AddHandler(Listener.MemberUpdate)
	dg.AddHandler(Listener.MemberUpdateTag)
	dg.AddHandler(Listener.MemberBanned)
	dg.AddHandler(Listener.MemberKicked)
	dg.AddHandler(Listener.MemberJoin)
	dg.AddHandler(Listener.MemberQuit)

	// Message Handler
	dg.AddHandler(Listener.MessageUpdate)
	dg.AddHandler(Listener.MessageCreate)
	dg.AddHandler(Listener.MessageDelete)
	dg.AddHandler(Listener.RolesReactionsAdd)
	dg.AddHandler(Listener.RolesReactionsRemove)

	// Command Manager
	dg.AddHandler(Command.CommandManager)

	dg.AddHandler(Listener.LevelsMessageCreate)
	dg.AddHandler(Listener.AntiBotListener)

	RegisterCommand(&Command.Role{})
	RegisterCommand(&Command.Levels{})
	RegisterCommand(&Command.Leaderboard{})
	RegisterCommand(&Command.RadioSearch{})

	err = dg.Open()
	if err != nil {
		fmt.Println("An error occured while attemping to start the discord bot:", err)
	}

	/*  ==== LAVALINK ==== */
	Radio.Link = disgolink.New(snowflake.MustParse(dg.State.User.ID), disgolink.WithListenerFunc(func(player disgolink.Player, event lavalink.Event) {
		switch e := event.(type) {
		case lavalink.TrackEndEvent:
			if e.Reason == lavalink.TrackEndReasonFinished {
				fmt.Println("Track finished playing. Attempting to play the next track...")
				go ConnectToRadioChannel(dg)
			}
		}
	}),
	)

	_, err = Radio.Link.AddNode(context.TODO(), disgolink.NodeConfig{
		Name:     "local-node",
		Address:  "127.0.0.1:2333",
		Password: os.Getenv("LAVALINK_PASSWORD"),
		Secure:   false,
	})
	if err != nil {
		fmt.Println("An error occured while attemping to connect to the Lavalink node:", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceServerUpdate) {
		Radio.Link.OnVoiceServerUpdate(context.TODO(), snowflake.MustParse(e.GuildID), e.Token, e.Endpoint)
	})

	// 2. Forwards your Bot's Voice Session ID to Lavalink
	dg.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		// Only forward updates for our own bot
		if e.UserID != s.State.User.ID {
			return
		}

		var channelID *snowflake.ID
		if e.ChannelID != "" {
			id := snowflake.MustParse(e.ChannelID)
			channelID = &id
		}

		Radio.Link.OnVoiceStateUpdate(context.TODO(), snowflake.MustParse(e.GuildID), channelID, e.SessionID)
	})

	/*  ==== LAVALINK ==== */

	fmt.Println("Ciao Kombucha, en ligne !")
	LoadCommands(dg)
	ConnectToRadioChannel(dg)

	/* ==== TEANO DAILY ==== */

	if os.Getenv("ENABLE_TEANO_DAILY") == "true" {
		go func() {
			sendDaily := func() {
				now := time.Now()
				first := time.Date(2026, 7, 29, 0, 0, 0, 0, now.Location())
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				count := int(today.Sub(first).Hours()/24) + 1
				if count < 1 {
					count = 1
				}

				msg := "Teano daily #" + fmt.Sprint(count)
				if _, err := dg.ChannelMessageSend("1442243971827896421", msg); err != nil {
					log.Println("failed to send Teano daily message:", err)
				}
			}

			sendDaily()

			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				sendDaily()
			}
		}()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()

}
