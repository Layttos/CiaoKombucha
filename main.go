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

	citations_query := `CREATE TABLE IF NOT EXISTS citations(
		id TEXT PRIMARY KEY,
		citation TEXT,
		author TEXT
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
	if _, err := Utils.DB.Exec(citations_query); err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("An error occured while attemping to initiate the Discord bot session:", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentsGuildBans | discordgo.IntentsGuildEmojis | discordgo.IntentsGuildInvites | discordgo.IntentsGuildWebhooks
	dg.State.TrackVoice = true

	// Member Handler
	dg.AddHandler(safeHandler(Listener.MemberUpdate))
	dg.AddHandler(safeHandler(Listener.MemberUpdateTag))
	dg.AddHandler(safeHandler(Listener.MemberBanned))
	dg.AddHandler(safeHandler(Listener.MemberKicked))
	dg.AddHandler(safeHandler(Listener.MemberJoin))
	dg.AddHandler(safeHandler(Listener.MemberQuit))

	// Message Handler
	dg.AddHandler(safeHandler(Listener.MessageUpdate))
	dg.AddHandler(safeHandler(Listener.MessageCreate))
	dg.AddHandler(safeHandler(Listener.MessageDelete))
	dg.AddHandler(safeHandler(Listener.MessageDeleteBulk))
	dg.AddHandler(safeHandler(Listener.RolesReactionsAdd))
	dg.AddHandler(safeHandler(Listener.RolesReactionsRemove))

	// Server structure Handler (rôles, salons, fils, paramètres, webhooks, invitations, emojis)
	dg.AddHandler(safeHandler(Listener.RoleCreated))
	dg.AddHandler(safeHandler(Listener.RoleUpdated))
	dg.AddHandler(safeHandler(Listener.RoleDeleted))
	dg.AddHandler(safeHandler(Listener.ChannelCreated))
	dg.AddHandler(safeHandler(Listener.ChannelUpdated))
	dg.AddHandler(safeHandler(Listener.ChannelDeleted))
	dg.AddHandler(safeHandler(Listener.ThreadCreated))
	dg.AddHandler(safeHandler(Listener.ThreadDeleted))
	dg.AddHandler(safeHandler(Listener.GuildUpdated))
	dg.AddHandler(safeHandler(Listener.GuildEmojisUpdated))
	dg.AddHandler(safeHandler(Listener.MemberUnbanned))
	dg.AddHandler(safeHandler(Listener.WebhooksUpdated))
	dg.AddHandler(safeHandler(Listener.InviteCreated))
	dg.AddHandler(safeHandler(Listener.InviteDeleted))

	// Command Manager
	dg.AddHandler(safeHandler(Command.CommandManager))

	dg.AddHandler(safeHandler(Listener.LevelsMessageCreate))
	dg.AddHandler(safeHandler(Listener.AntiBotListener))
	dg.AddHandler(safeHandler(Listener.EmailBotJoin))

	RegisterCommand(&Command.Role{})
	RegisterCommand(&Command.Levels{})
	RegisterCommand(&Command.Leaderboard{})
	RegisterCommand(&Command.Snipe{})
	RegisterCommand(&Command.Say{})
	if os.Getenv("ENABLE_RADIO") == "true" {
		RegisterCommand(&Command.RadioSearch{})
		RegisterCommand(&Command.Skip{})
		RegisterCommand(&Command.Play{})
	}

	if err := openSession(dg); err != nil {
		log.Fatal("An error occured while attemping to start the discord bot: ", err)
	}

	if dg.State == nil || dg.State.User == nil {
		log.Fatal("Discord n'a pas renvoyé de paquet READY exploitable : identité du bot inconnue.")
	}

	/*  ==== LAVALINK ==== */
	if os.Getenv("ENABLE_RADIO") == "true" {
		Radio.Link = disgolink.New(
			snowflake.MustParse(dg.State.User.ID),
			disgolink.WithListenerFunc(Radio.LavalinkEventHandler(dg)),
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

		dg.AddHandler(safeHandler(func(s *discordgo.Session, e *discordgo.VoiceServerUpdate) {
			guildID, err := snowflake.Parse(e.GuildID)
			if Radio.Link == nil || err != nil {
				return
			}

			Radio.Link.OnVoiceServerUpdate(context.TODO(), guildID, e.Token, e.Endpoint)
		}))

		// 2. Forwards your Bot's Voice Session ID to Lavalink
		dg.AddHandler(safeHandler(func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
			// Only forward updates for our own bot
			if Radio.Link == nil || s.State == nil || s.State.User == nil || e.UserID != s.State.User.ID {
				return
			}

			guildID, err := snowflake.Parse(e.GuildID)
			if err != nil {
				return
			}

			var channelID *snowflake.ID
			if e.ChannelID != "" {
				id, err := snowflake.Parse(e.ChannelID)
				if err != nil {
					return
				}
				channelID = &id
			}

			Radio.Link.OnVoiceStateUpdate(context.TODO(), guildID, channelID, e.SessionID)
		}))
	}
	/*  ==== LAVALINK ==== */

	fmt.Println("Ciao Kombucha, en ligne !")
	LoadCommands(dg)
	if os.Getenv("ENABLE_RADIO") == "true" {
		ConnectToRadioChannel(dg)
		Radio.StartWatchdog(dg)
	}

	/* ==== TEANO DAILY ==== */

	if os.Getenv("ENABLE_TEANO_DAILY") == "true" {
		go func() {
			sendDaily := func() {
				defer Utils.RecoverPanic("le Teano daily")

				now := time.Now()
				first := time.Date(2026, 7, 29, 0, 0, 0, 0, now.Location())
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				count := int(today.Sub(first).Hours()/24) + 1
				if count < 1 {
					count = 1
				}

				msg := "Teano daily #" + fmt.Sprint(count)
				if _, err := dg.ChannelMessageSend("1442243974311182358", msg); err != nil {
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

	startGatewayWatchdog(dg)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()

}
