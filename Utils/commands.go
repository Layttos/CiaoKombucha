package Utils

import "github.com/bwmarrin/discordgo"

type Command interface {
	Name() string
	Description() string

	Permissions() *int64
	Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool
}

var Commands []Command
