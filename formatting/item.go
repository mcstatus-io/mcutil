package formatting

import (
	"fmt"
	"html"
	"reflect"
	"strings"

	"github.com/mcstatus-io/mcutil/v4/formatting/colors"
	"github.com/mcstatus-io/mcutil/v4/formatting/decorators"
)

// Item is a single formatting item that contains a color and any optional formatting controls.
type Item struct {
	Text       string                 `json:"text"`
	Color      *colors.Color          `json:"color,omitempty"`
	Decorators []decorators.Decorator `json:"decorators"`
}

// Raw returns the Minecraft encoding of the formatting (§ + formatting codes/color + text).
func (i Item) Raw() (result string) {
	if i.Color != nil {
		result += i.Color.ToRaw()
	}

	for _, decorator := range i.Decorators {
		result += decorator.ToRaw()
	}

	result += i.Text

	return
}

// Clean returns the text only of the formatting item.
func (i Item) Clean() string {
	return i.Text
}

// HTML returns the HTML representation of the format item.
// A <span> tag is used as well as the style and class attributes.
// Obfuscation is encoded as a class with the name 'minecraft-format-obfuscated'.
func (i Item) HTML() string {
	classes := make([]string, 0)

	styles := map[string]string{}

	if i.Color != nil {
		styles["color"] = i.Color.ToHex()
	}

	for _, decorator := range i.Decorators {
		switch decorator {
		case decorators.Obfuscated:
			{
				classes = append(classes, "minecraft-format-obfuscated")

				break
			}
		case decorators.Bold:
			{
				styles["font-weight"] = "bold"

				break
			}
		case decorators.Strikethrough:
			{
				if _, ok := styles["text-decoration"]; ok {
					styles["text-decoration"] += " "
				}

				styles["text-decoration"] += "line-through"

				break
			}
		case decorators.Underlined:
			{
				if _, ok := styles["text-decoration"]; ok {
					styles["text-decoration"] += " "
				}

				styles["text-decoration"] += "underline"

				break
			}
		case decorators.Italic:
			{
				styles["font-style"] = "italic"

				break
			}
		}
	}

	var (
		rawClass string = ""
		rawStyle string = ""
	)

	if len(classes) > 0 {
		rawClass = fmt.Sprintf(" class=\"%s\"", strings.Join(classes, " "))
	}

	if len(styles) > 0 {
		styleData := ""

		for k, v := range styles {
			styleData += fmt.Sprintf("%s: %s;", k, v)
		}

		rawStyle = fmt.Sprintf(" style=\"%s\"", styleData)
	}

	return fmt.Sprintf("<span%s%s>%s</span>", rawClass, rawStyle, html.EscapeString(i.Text))
}

// IsSameAs returns whether the formatting is identical to another
// item.
func (i Item) IsSameAs(j Item) bool {
	if !reflect.DeepEqual(i.Color, j.Color) || len(i.Decorators) != len(j.Decorators) {
		return false
	}

	for _, a := range i.Decorators {
		if !contains(j.Decorators, a) {
			return false
		}
	}

	for _, a := range j.Decorators {
		if !contains(i.Decorators, a) {
			return false
		}
	}

	return true
}

func toRaw(tree []Item) (result string) {
	for _, v := range tree {
		result += v.Raw()
	}

	return
}

func toClean(tree []Item) (result string) {
	for _, v := range tree {
		result += v.Clean()
	}

	return
}

func toHTML(tree []Item) (result string) {
	result = "<span>"

	for _, v := range tree {
		result += v.HTML()
	}

	result += "</span>"

	return
}

func contains[T comparable](arr []T, v T) bool {
	for _, a := range arr {
		if a == v {
			return true
		}
	}

	return false
}
