package Command

import (
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func CommandManager(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, cmd := range Utils.Commands {
		if i.ApplicationCommandData().Name == cmd.Name() {
			cmd.Execute(s, i)
		}
	}
}
