package Radio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/delucks/go-subsonic"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

var Link disgolink.Client

type Navidrome struct {
	Client *subsonic.Client
}

func ConnectToRadioChannel(s *discordgo.Session) error {
	rawUrl := os.Getenv("NAVIDROME_URL")
	rawUser := os.Getenv("NAVIDROME_USER")
	password := os.Getenv("NAVIDROME_PASSWORD")
	if rawUrl == "" || rawUser == "" || password == "" {
		fmt.Println("One or more required environment variables are not set for the Navidrome client.")
		return fmt.Errorf("missing environment variables")
	}

	client := &subsonic.Client{
		Client:     &http.Client{},
		ClientName: "CiaoKombucha",
		BaseUrl:    strings.TrimSpace(rawUrl),
		User:       rawUser,
	}

	if err := client.Authenticate(password); err != nil {
		fmt.Println("Authentication failed:", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	if !client.Ping() {
		return fmt.Errorf("Failed to ping the Navidrome server. Please check the URL and credentials stored in the .env file.")
	}

	fmt.Println("Attempting to join the radio channel, " + os.Getenv("RADIO_CHANNEL_ID"))
	err := s.ChannelVoiceJoinManual(os.Getenv("GUILD_ID"), os.Getenv("RADIO_CHANNEL_ID"), false, false)
	if err != nil {
		return fmt.Errorf("An error occurred while attempting to join the radio channel: %w", err)
	}

	return ChangeCurrentTrack("")
}

func ChangeCurrentTrack(query string) error {
	rawUrl := os.Getenv("NAVIDROME_URL")
	rawUser := os.Getenv("NAVIDROME_USER")
	password := os.Getenv("NAVIDROME_PASSWORD")

	client := &subsonic.Client{
		Client:     &http.Client{},
		ClientName: "CiaoKombucha",
		BaseUrl:    strings.TrimSpace(rawUrl),
		User:       rawUser,
	}

	if err := client.Authenticate(password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	var targetSong *subsonic.Child

	if strings.TrimSpace(query) != "" {
		parameters := map[string]string{"songCount": "1"}
		searchResult, err := client.Search3(query, parameters)

		if err != nil || searchResult == nil || len(searchResult.Song) == 0 {
			return fmt.Errorf("Search failed or returned no results for query: %s, Error: %w", query, err)
		} else {
			targetSong = searchResult.Song[0]
			fmt.Println("Result found:", targetSong.Title, "by", targetSong.Artist)
		}
	}

	if targetSong == nil || targetSong.ID == "" {
		songs, err := client.GetRandomSongs(map[string]string{"size": "1"})
		if err != nil || len(songs) == 0 {
			return fmt.Errorf("failed to get random songs: %w", err)
		}

		targetSong = songs[0]

		fmt.Println("Now playing random track:", targetSong.Title, "by", targetSong.Artist)
	}

	encodedPassword := url.QueryEscape(password)
	streamURL := fmt.Sprintf("%s/rest/stream?id=%s&u=%s&p=%s&v=1.16.1&c=CiaoKombucha",
		client.BaseUrl,
		targetSong.ID,
		rawUser,
		encodedPassword,
	)

	player := Link.Player(snowflake.MustParse(os.Getenv("GUILD_ID")))

	result, err := Link.BestNode().LoadTracks(context.TODO(), streamURL)
	if err != nil {
		return fmt.Errorf("lavalink connection error: %w", err)
	}

	switch data := result.Data.(type) {
	case lavalink.Track:
		err := player.Update(context.TODO(), lavalink.WithTrack(data), lavalink.WithPaused(false))
		if err != nil {
			return fmt.Errorf("Error playing track via Lavalink: %w", err)
		}
		fmt.Println("Track now playing!")

	case lavalink.Exception:
		fmt.Println("Lavalink error loading track:", data.Message)
		return fmt.Errorf("An error occured while attemping to load a track via lavalink: %s", data.Message)

	default:
		if result.LoadType == lavalink.LoadTypeEmpty {
			return fmt.Errorf("No matches found for the stream URL: %s", streamURL)
		}
		return fmt.Errorf("An unexpected error has occured while attemping to load the track via Lavalink: %s", result.LoadType)
	}

	return nil
}
