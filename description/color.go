package description

type Color rune

var (
	Black        Color = '0'
	DarkBlue     Color = '1'
	DarkGreen    Color = '2'
	DarkAqua     Color = '3'
	DarkRed      Color = '4'
	DarkPurple   Color = '5'
	Gold         Color = '6'
	Gray         Color = '7'
	DarkGray     Color = '8'
	Blue         Color = '9'
	Green        Color = 'a'
	Aqua         Color = 'b'
	Red          Color = 'c'
	LightPurple  Color = 'd'
	Yellow       Color = 'e'
	White        Color = 'f'
	MinecoinGold Color = 'g'
)

func ParseColor(value interface{}) Color {
	switch value {
	case "0", "black", Black:
		return Black
	case "1", "dark_blue", DarkBlue:
		return DarkBlue
	case "2", "dark_green", DarkGreen:
		return DarkGreen
	case "3", "dark_aqua", DarkAqua:
		return DarkAqua
	case "4", "dark_red", DarkRed:
		return DarkRed
	case "5", "dark_purple", DarkPurple:
		return DarkPurple
	case "6", "gold", Gold:
		return Gold
	case "7", "gray", Gray:
		return Gray
	case "8", "dark_gray", DarkGray:
		return DarkGray
	case "9", "blue", Blue:
		return Blue
	case "a", "green", Green:
		return Green
	case "b", "aqua", Aqua:
		return Aqua
	case "c", "red", Red:
		return Red
	case "d", "light_purple", LightPurple:
		return LightPurple
	case "e", "yellow", Yellow:
		return Yellow
	case "f", "white", White:
		return White
	case "g", "minecoin_gold", MinecoinGold:
		return MinecoinGold
	default:
		return White
	}
}

func (c Color) ToRaw() string {
	return "\u00A7" + string(c)
}

func (c Color) ToHex() string {
	switch c {
	case Black:
		return "#000000"
	case DarkBlue:
		return "#0000aa"
	case DarkGreen:
		return "#00aa00"
	case DarkAqua:
		return "#00aaaa"
	case DarkRed:
		return "#aa0000"
	case DarkPurple:
		return "#aa00aa"
	case Gold:
		return "#ffaa00"
	case Gray:
		return "#aaaaaa"
	case DarkGray:
		return "#555555"
	case Blue:
		return "#5555ff"
	case Green:
		return "#55ff55"
	case Aqua:
		return "#55ffff"
	case Red:
		return "#ff5555"
	case LightPurple:
		return "#ff55ff"
	case Yellow:
		return "#ffff55"
	case White:
		return "#ffffff"
	case MinecoinGold:
		return "#ddd605"
	default:
		return "#ffffff"
	}
}
