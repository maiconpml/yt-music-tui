package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	BorderColor = lipgloss.Color("62")

	DocStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Background(BorderColor).Foreground(lipgloss.Color("230"))
	BaseItemStyle     = lipgloss.NewStyle().PaddingLeft(2)

	NameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	DimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 4)
)

func ContainerFrameWidth() int {
	fw, _ := DocStyle.GetFrameSize()
	return fw
}

func ContainerFrameHeight() int {
	return 4 // topLine (1) + paddingTop (1) + paddingBottom (1) + borderBottom (1)
}

func Truncate(s string, max int) string {
	if max <= 3 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

func RenderContainer(width int, content string, titles ...string) string {
	if width == 0 {
		return ""
	}

	borderStyle := lipgloss.NewStyle().Foreground(BorderColor)
	border := lipgloss.RoundedBorder()

	titleText := ""
	for _, t := range titles {
		if t != "" {
			titleText += borderStyle.Render(border.Top) + " " + t + " "
		}
	}

	tl := borderStyle.Render(border.TopLeft)
	tr := borderStyle.Render(border.TopRight)
	t := borderStyle.Render(border.Top)

	topBarWidth := width - lipgloss.Width(titleText) - 3
	topBarWidth = max(topBarWidth, 0)

	topLine := tl + t + titleText + strings.Repeat(t, topBarWidth) + tr

	fw := ContainerFrameWidth()
	bodyContent := lipgloss.NewStyle().Width(width - fw).Render(content)

	body := DocStyle.
		UnsetBorderTop().
		PaddingTop(1).
		Render(bodyContent)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, body)
}
