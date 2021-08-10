package eod

import (
	"fmt"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/bwmarrin/discordgo"
)

var noModCmds = map[string]types.Empty{
	"suggest":      {},
	"mark":         {},
	"image":        {},
	"inv":          {},
	"lb":           {},
	"addcat":       {},
	"cat":          {},
	"hint":         {},
	"stats":        {},
	"idea":         {},
	"help":         {},
	"rmcat":        {},
	"catimg":       {},
	"downloadinv":  {},
	"elemsort":     {},
	"breakdown":    {},
	"catbreakdown": {},
}

func (b *EoD) canRunCmd(cmd *discordgo.InteractionCreate) (bool, string) {
	resp := cmd.ApplicationCommandData()

	// Check if mod is not required
	_, exists := noModCmds[resp.Name]
	if exists {
		return true, ""
	}

	// Check if is mod
	ismod, err := b.isMod(cmd.Member.User.ID, cmd.GuildID, b.newMsgSlash(cmd))
	if err != nil {
		return false, err.Error()
	}
	if ismod {
		return true, ""
	}

	// Get dat because everything after will require it
	lock.RLock()
	dat, exists := b.dat[cmd.GuildID]
	lock.RUnlock()
	falseMsg := "You need to have permission `Administrator` or have role <@&" + dat.ModRole + ">!"
	if !exists {
		return false, falseMsg
	}

	// If command is path or catpath, check if has element/all elements in cat
	// path
	if resp.Name == "path" || resp.Name == "graph" {
		dat.Lock.RLock()
		inv, exists := dat.InvCache[cmd.Member.User.ID]
		dat.Lock.RUnlock()
		if !exists {
			return false, "You don't have an inventory!"
		}

		name := strings.ToLower(resp.Options[0].StringValue())
		dat.Lock.RLock()
		el, exists := dat.ElemCache[name]
		dat.Lock.RUnlock()
		if !exists {
			return true, "" // If the element doesn't exist, the cat command will tell the user it doesn't exist
		}

		_, exists = inv[name]
		if !exists {
			return false, fmt.Sprintf("You must have element **%s** to get it's path!", el.Name)
		}
		return true, ""
	}

	// catpath
	if resp.Name == "catpath" || resp.Name == "catgraph" {
		dat.Lock.RLock()
		inv, exists := dat.InvCache[cmd.Member.User.ID]
		dat.Lock.RUnlock()
		if !exists {
			return false, "You don't have an inventory!"
		}
		cat, exists := dat.CatCache[strings.ToLower(resp.Options[0].StringValue())]
		if !exists {
			return true, "" // If the category doesn't exist, the cat command will tell the user it doesn't exist
		}

		// Check if user has all elements in category
		for elem := range cat.Elements {
			_, exists = inv[strings.ToLower(elem)]
			if !exists {
				return false, fmt.Sprintf("You must have all elements in category **%s** to get its path!", cat.Name)
			}
		}

		return true, ""
	}

	return false, falseMsg
}