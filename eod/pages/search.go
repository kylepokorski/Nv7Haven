package pages

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/Nv7-Github/sevcord/v2"
	"github.com/dustin/go-humanize"
)

func SearchInputs(args []any) (string, int, int) {
	sort := ""
	if args[1] != nil {
		sort = args[1].(string)
	}
	postfix := false
	postfixval := 0

	if args[2] != nil {
		postfix = args[2].(bool)
	}
	if postfix {
		postfixval = 1
	} else {
		postfixval = 0
	}
	page := -1
	if len(args) > 3 && args[3] != nil {
		page = int(args[3].(int64)) - 2
	}
	return sort, postfixval, page
}
func (p *Pages) SearchPrefix(c sevcord.Ctx, args []any) {
	c.Acknowledge()
	sort, postfixval, page := SearchInputs(args)
	p.SearchHandler(c, fmt.Sprintf("next|%s|%s|%d|%d|%s|prefix", c.Author().User.ID, sort, postfixval, page, args[0].(string)))
}
func (p *Pages) SearchRegex(c sevcord.Ctx, args []any) {
	c.Acknowledge()
	sort, postfixval, page := SearchInputs(args)
	p.SearchHandler(c, fmt.Sprintf("next|%s|%s|%d|%d|%s|regex", c.Author().User.ID, sort, postfixval, page, args[0].(string)))
}

// Format: prevnext|user|sort|postfix|page|searchquery|searchtype
func (p *Pages) SearchHandler(c sevcord.Ctx, params string) {

	parts := strings.Split(params, "|")
	length := p.base.PageLength(c)
	cnt := 0
	cond := "ILIKE $2||'%'"
	sorttype := "similarity(name,$2) DESC"
	switch parts[6] {
	case "prefix":
		cond = "ILIKE $2||'%'"
		sorttype = "similarity(name,$2) DESC"
	case "regex":
		cond = "~ $2"
		sorttype = "id"
	}
	err := p.db.QueryRow("SELECT COUNT(*) from elements WHERE guild=$1 AND name "+cond, c.Guild(), parts[5]).Scan(&cnt)
	if err != nil {
		return
	}

	pagecnt := int(math.Ceil(float64(cnt) / float64(length)))

	// Apply page
	page, _ := strconv.Atoi(parts[4])
	page = ApplyPage(parts[0], page, pagecnt)

	// Get values
	var items []struct {
		Name    string `db:"name"`
		Cont    bool   `db:"cont"`
		Postfix string `db:"postfix"`
	}

	postfix := false
	if parts[3] == "1" {
		postfix = true
	} else {
		postfix = false
	}
	//false if not valid in DB
	postfixable := parts[2] != "length" && parts[2] != "found" && parts[2] != ""
	postfixadd := ""

	if postfix && postfixable {
		postfixadd = "," + parts[2] + " postfix"
	} else {
		if parts[2] != "" {
			sorttype = types.SortSql[parts[2]]
		}

	}

	querystr := fmt.Sprintf(`SELECT name,id=ANY(SELECT UNNEST(inv) FROM inventories WHERE guild=$1 AND "user"=$5) cont %s FROM elements  WHERE name %s AND guild=$1 ORDER BY %s LIMIT $3 OFFSET $4`, postfixadd, cond, sorttype)

	err = p.db.Select(&items, querystr, c.Guild(), parts[5], length, length*page, parts[1])

	if err != nil {
		p.base.Error(c, err)
		return
	}
	desc := &strings.Builder{}
	for _, v := range items {
		if v.Cont {
			fmt.Fprintf(desc, "%s %s", v.Name, types.Check)
		} else {
			fmt.Fprintf(desc, "%s %s", v.Name, types.NoCheck)
		}
		if postfix && parts[2] != "found" {
			desc.WriteString(p.PrintPostfix(parts[2], v.Name, v.Postfix))

		}
		desc.WriteString("\n")

	}
	color := 10181046 //Purple
	embed := sevcord.NewEmbed().
		Title(fmt.Sprintf(`Found %s results for "%s"`, humanize.Comma(int64(cnt)), parts[5])).
		Description(desc.String()).
		Footer(fmt.Sprintf("Page %d/%d", page+1, pagecnt), "").
		Color(color)
	c.Respond(sevcord.NewMessage("").
		AddEmbed(embed).
		AddComponentRow(PageSwitchBtns("search", fmt.Sprintf("%s|%s|%s|%d|%s|%s", parts[1], parts[2], parts[3], page, parts[5], parts[6]))...),
	)
}
