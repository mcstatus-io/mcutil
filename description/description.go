package description

import (
	"fmt"
	"html"
	"strings"

	"github.com/fatih/color"
)

var (
	formattingColorCodeLookupTable = map[rune]string{
		'0': "black",
		'1': "dark_blue",
		'2': "dark_green",
		'3': "dark_aqua",
		'4': "dark_red",
		'5': "dark_purple",
		'6': "gold",
		'7': "gray",
		'8': "dark_gray",
		'9': "blue",
		'a': "green",
		'b': "aqua",
		'c': "red",
		'd': "light_purple",
		'e': "yellow",
		'f': "white",
		'g': "minecoin_gold",
	}
	colorNameLookupTable = map[string]rune{
		"black":         '0',
		"dark_blue":     '1',
		"dark_green":    '2',
		"dark_aqua":     '3',
		"dark_red":      '4',
		"dark_purple":   '5',
		"gold":          '6',
		"gray":          '7',
		"dark_gray":     '8',
		"blue":          '9',
		"green":         'a',
		"aqua":          'b',
		"red":           'c',
		"light_purple":  'd',
		"yellow":        'e',
		"white":         'f',
		"minecoin_gold": 'g',
	}
	htmlColorLookupTable = map[string]string{
		"black":         "#000000",
		"dark_blue":     "#0000aa",
		"dark_green":    "#00aa00",
		"dark_aqua":     "#00aaaa",
		"dark_red":      "#aa0000",
		"dark_purple":   "#aa00aa",
		"gold":          "#ffaa00",
		"gray":          "#aaaaaa",
		"dark_gray":     "#555555",
		"blue":          "#5555ff",
		"green":         "#55ff55",
		"aqua":          "#55ffff",
		"red":           "#ff5555",
		"light_purple":  "#ff55ff",
		"yellow":        "#ffff55",
		"white":         "#ffffff",
		"minecoin_gold": "#ddd605",
	}
	ansiColorLookupTable = map[string]color.Attribute{
		"black":         color.FgBlack,
		"dark_blue":     color.FgBlue,
		"dark_green":    color.FgGreen,
		"dark_aqua":     color.FgCyan,
		"dark_red":      color.FgRed,
		"dark_purple":   color.FgMagenta,
		"gold":          color.FgYellow,
		"gray":          color.FgHiBlack,
		"dark_gray":     color.FgHiBlack,
		"blue":          color.FgHiBlue,
		"green":         color.FgHiGreen,
		"aqua":          color.FgHiCyan,
		"red":           color.FgHiRed,
		"light_purple":  color.FgHiMagenta,
		"yellow":        color.FgHiYellow,
		"white":         color.FgHiWhite,
		"minecoin_gold": color.FgHiYellow,
	}
)

type FormatItem struct {
	Text          string `json:"text"`
	Color         string `json:"color"`
	Obfuscated    bool   `json:"obfuscated"`
	Bold          bool   `json:"bold"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Italic        bool   `json:"italic"`
}

type MOTD struct {
	Tree  []FormatItem `json:"-"`
	ANSI  string       `json:"-"`
	Raw   string       `json:"raw"`
	Clean string       `json:"clean"`
	HTML  string       `json:"html"`
}

func (m MOTD) String() string {
	return m.Clean
}

func ParseMOTD(desc interface{}) (*MOTD, error) {
	var tree []FormatItem

	switch v := desc.(type) {
	case string:
		{
			t, err := parseString(v)

			if err != nil {
				return nil, err
			}

			tree = t

			break
		}
	case map[string]interface{}:
		{
			t, err := parseString(parseChatObject(v))

			if err != nil {
				return nil, err
			}

			tree = t

			break
		}
	default:
		return nil, fmt.Errorf("unknown description type: %T", desc)
	}

	return &MOTD{
		Tree:  tree,
		ANSI:  toANSI(tree),
		Raw:   toRaw(tree),
		Clean: toClean(tree),
		HTML:  toHTML(tree),
	}, nil
}

func toRaw(tree []FormatItem) string {
	result := ""

	for _, v := range tree {
		if v.Color != "white" {
			colorCode, ok := colorNameLookupTable[v.Color]

			if ok {
				result += "\u00A7" + string(colorCode)
			}
		}

		if v.Obfuscated {
			result += "\u00A7k"
		}

		if v.Bold {
			result += "\u00A7l"
		}

		if v.Strikethrough {
			result += "\u00A7m"
		}

		if v.Underline {
			result += "\u00A7n"
		}

		if v.Italic {
			result += "\u00A7o"
		}

		result += v.Text
	}

	return result
}

func toClean(tree []FormatItem) (res string) {
	for _, v := range tree {
		res += v.Text
	}

	return
}

func toHTML(tree []FormatItem) (res string) {
	res = "<span>"

	for _, v := range tree {
		classes := make([]string, 0)
		styles := make(map[string]string)

		color, ok := htmlColorLookupTable[v.Color]

		if ok {
			styles["color"] = color
		}

		if v.Obfuscated {
			classes = append(classes, "minecraft-format-obfuscated")
		}

		if v.Bold {
			styles["font-weight"] = "bold"
		}

		if v.Strikethrough {
			if _, ok = styles["text-decoration"]; ok {
				styles["text-decoration"] += " "
			}

			styles["text-decoration"] += "line-through"
		}

		if v.Underline {
			if _, ok = styles["text-decoration"]; ok {
				styles["text-decoration"] += " "
			}

			styles["text-decoration"] += "underline"
		}

		if v.Italic {
			styles["font-style"] = "italic"
		}

		res += "<span"

		if len(classes) > 0 {
			res += " class=\""

			for _, v := range classes {
				res += v
			}

			res += "\""
		}

		if len(styles) > 0 {
			res += " style=\""

			keys := make([]string, 0, len(styles))

			for k := range styles {
				keys = append(keys, k)
			}

			for i, l := 0, len(keys); i < l; i++ {
				key := keys[i]
				value := styles[key]

				res += fmt.Sprintf("%s: %s;", key, value)

				if i+1 != l {
					res += " "
				}
			}

			res += "\""
		}

		res += fmt.Sprintf(">%s</span>", html.EscapeString(v.Text))
	}

	res += "</span>"

	return
}

func toANSI(tree []FormatItem) (res string) {
	for _, v := range tree {
		attr := make([]color.Attribute, 0)

		if color, ok := ansiColorLookupTable[v.Color]; ok {
			attr = append(attr, color)
		}

		if v.Bold {
			attr = append(attr, color.Bold)
		}

		if v.Strikethrough {
			attr = append(attr, color.CrossedOut)
		}

		if v.Underline {
			attr = append(attr, color.Underline)
		}

		if v.Italic {
			attr = append(attr, color.Italic)
		}

		res += color.New(attr...).Sprintf(v.Text)
	}

	return
}

func parseChatObject(m map[string]interface{}) (res string) {
	color, ok := m["color"].(string)

	if ok {
		code, ok := colorNameLookupTable[color]

		if ok {
			res += "\u00A7" + string(code)
		}
	}

	bold, ok := m["bold"].(string)

	if ok && bold == "true" {
		res += "\u00A7l"
	}

	italic, ok := m["italic"].(string)

	if ok && italic == "true" {
		res += "\u00A7o"
	}

	underline, ok := m["underlined"].(string)

	if ok && underline == "true" {
		res += "\u00A7n"
	}

	strikethrough, ok := m["strikethrough"].(string)

	if ok && strikethrough == "true" {
		res += "\u00A7m"
	}

	obfuscated, ok := m["obfuscated"].(string)

	if ok && obfuscated == "true" {
		res += "\u00A7k"
	}

	text, ok := m["text"].(string)

	if ok {
		res += text
	}

	extra, ok := m["extra"].([]interface{})

	if ok {
		for _, v := range extra {
			v2, ok := v.(map[string]interface{})

			if ok {
				res += parseChatObject(v2)
			}
		}
	}

	return
}

func parseString(s string) ([]FormatItem, error) {
	tree := make([]FormatItem, 0)

	item := FormatItem{
		Text:  "",
		Color: "white",
	}

	r := strings.NewReader(s)

	for r.Len() > 0 {
		char, n, err := r.ReadRune()

		if err != nil {
			return nil, err
		}

		if n < 1 {
			break
		}

		if char == '\n' {
			tree = append(tree, item)

			item = FormatItem{
				Text:  "\n",
				Color: "white",
			}

			continue
		}

		if char != '\u00A7' {
			item.Text += string(char)

			continue
		}

		code, n, err := r.ReadRune()

		if err != nil {
			return nil, err
		}

		if n < 1 {
			break
		}

		// Color code
		{
			name, ok := formattingColorCodeLookupTable[code]

			if ok {
				if item.Obfuscated || item.Bold || item.Strikethrough || item.Underline || item.Italic || name != item.Color {
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item = FormatItem{
						Text:  "",
						Color: name,
					}
				} else {
					item.Color = name
				}

				continue
			}
		}

		// Formatting code
		{
			switch code {
			case 'k':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Obfuscated = true
				}
			case 'l':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Bold = true
				}
			case 'm':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Strikethrough = true
				}
			case 'n':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Underline = true
				}
			case 'o':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Italic = true
				}
			case 'r':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item = FormatItem{
						Text:  "",
						Color: "white",
					}
				}
			}
		}
	}

	tree = append(tree, item)

	return tree, nil
}
