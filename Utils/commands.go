package Utils

import "github.com/bwmarrin/discordgo"

type Command interface {
	Name() string
	Description() string

	Permissions() *int64
	Options() []*discordgo.ApplicationCommandOption
	Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool
}

var Commands []Command
