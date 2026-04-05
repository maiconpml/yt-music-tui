package home

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/tui/styles"
)

type Model struct {
	width  int
	height int
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

var (
	styleRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	styleDefault = lipgloss.NewStyle().Foreground(lipgloss.Color("254"))
)

var logoHey = styleRed.Render(`██╗  ██╗███████╗██╗   ██╗  
██║  ██║██╔════╝╚██╗ ██╔╝  
███████║█████╗   ╚████╔╝   
██╔══██║██╔══╝    ╚██╔╝    
██║  ██║███████╗   ██║     
╚═╝  ╚═╝╚══════╝   ╚═╝     
`)

var logoH = styleRed.Render(`██╗  ██╗
██║  ██║
███████║
██╔══██║
██║  ██║
╚═╝  ╚═╝
`)

var logoH1 = styleRed.Render(`██╗  
██║  
█████
██╔══
██║  
╚═╝  
`)

var logoListen = styleDefault.Render(`██╗     ██╗███████╗████████╗███████╗███╗   ██╗██╗
██║     ██║██╔════╝╚══██╔══╝██╔════╝████╗  ██║██║
██║     ██║███████╗   ██║   █████╗  ██╔██╗ ██║██║
██║     ██║╚════██║   ██║   ██╔══╝  ██║╚██╗██║╚═╝
███████╗██║███████║   ██║   ███████╗██║ ╚████║██╗
╚══════╝╚═╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═══╝╚═╝
`)

var logoLsn = styleDefault.Render(`  ██╗     ███████╗███╗   ██╗██╗
  ██║     ██╔════╝████╗  ██║██║
  ██║     ███████╗██╔██╗ ██║██║
  ██║     ╚════██║██║╚██╗██║╚═╝
  ███████╗███████║██║ ╚████║██╗
  ╚══════╝╚══════╝╚═╝  ╚═══╝╚═╝
`)

var logoL = styleDefault.Render(`██╗     ██╗
██║     ██║
██║     ██║
██║     ╚═╝
███████╗██╗
╚══════╝╚═╝
`)

var logoExc = styleDefault.Render(`██╗
██║
██║
╚═╝
██╗
╚═╝
`)

var (
	logoXBig   = lipgloss.JoinHorizontal(lipgloss.Center, logoHey, logoListen)
	logoBig    = lipgloss.JoinHorizontal(lipgloss.Center, logoHey, logoLsn)
	logoMedium = lipgloss.JoinHorizontal(lipgloss.Center, logoH, logoLsn)
	logoSmall  = lipgloss.JoinHorizontal(lipgloss.Center, logoH, logoL)
	logoXSmall = lipgloss.JoinHorizontal(lipgloss.Center, logoH1, logoExc)
	xBigSize   = lipgloss.Width(logoXBig)
	bigSize    = lipgloss.Width(logoBig)
	medSize    = lipgloss.Width(logoMedium)
	smallSize  = lipgloss.Width(logoSmall)
	xSmallSize = lipgloss.Width(logoXSmall)
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	logo := logoXBig
	w := m.width - 6
	if w < smallSize {
		logo = logoXSmall
	} else if w < medSize {
		logo = logoSmall
	} else if w < bigSize {
		logo = logoMedium
	} else if w < xBigSize {
		logo = logoBig
	}
	logoView := lipgloss.Place(m.width-4, m.height-2, lipgloss.Center, lipgloss.Center, logo)

	return styles.RenderContainer(m.width, logoView, "")
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}
