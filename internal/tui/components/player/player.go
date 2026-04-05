package player

import (
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/audio"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/internal/ytdlp"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

type PlayerTickMsg time.Time

type TogglePauseMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return PlayerTickMsg(t)
	})
}

type QueueLoadedMsg struct {
	Tracks       []*goytmusic.Track
	CurTrack     int
	Continuation string
}

type PrefetchCompleteMsg struct {
	Index int
	Err   error
}

type StreamReadyMsg struct {
	Data io.Reader
	Err  error
}

type (
	StreamStartedMsg struct{}
	StreamErrorMsg   struct {
		Err error
	}
)

type (
	NextTrackMsg struct{}
	PrevTrackMsg struct{}
)

type playerStatus int

const (
	standby playerStatus = iota
	playing
	paused
	downloading
)

type Model struct {
	status           playerStatus
	tracks           []*goytmusic.Track
	spinner          spinner.Model
	curTrack         int
	continuation     string
	continuationType int
	width            int
	progress         progress.Model
	position         float64
	duration         float64
	err              error
}

func New() Model {
	prog := progress.New(progress.WithFillCharacters('󱘹', '󰍴'), progress.WithoutPercentage())
	return Model{
		status:   standby,
		curTrack: -1,
		progress: prog,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Points)),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.spinner.Tick)
}

func formatDuration(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func parseDuration(durStr string) float64 {
	if durStr == "" {
		return 0
	}
	parts := strings.Split(durStr, ":")
	var totalSeconds float64
	for i, part := range parts {
		val, _ := strconv.Atoi(part)
		multiplier := 1.0
		power := len(parts) - 1 - i
		if power == 1 {
			multiplier = 60
		} else if power == 2 {
			multiplier = 3600
		}
		totalSeconds += float64(val) * multiplier
	}
	return totalSeconds
}

func (m *Model) prefetchTrack(index int) tea.Cmd {
	if index < 0 || index >= len(m.tracks) {
		return nil
	}

	tr := m.tracks[index]
	if tr.VideoID == nil {
		return nil
	}
	return func() tea.Msg {
		slog.Info("prefetching track", "name", tr.Name)
		_, err := ytdlp.GetAudioPath(*tr.VideoID)
		if err != nil {
			slog.Error("error prefetching track", "error", err.Error())
		}
		slog.Info("track prefetched sucessfully")
		return PrefetchCompleteMsg{Index: index, Err: err}
	}
}

func (m *Model) playTrack(index int) tea.Cmd {
	if index < 0 || index >= len(m.tracks) {
		return nil
	}
	m.curTrack = index
	m.status = downloading
	m.position = 0
	m.err = nil

	tr := m.tracks[m.curTrack]
	m.duration = parseDuration(tr.Duration)

	if tr.VideoID == nil {
		m.err = fmt.Errorf("track without videoID")
		m.status = standby
		return nil
	}

	videoID := *tr.VideoID
	return func() tea.Msg {
		slog.Info("configuring stream for track", "name", tr.Name)
		data, err := ytdlp.StreamAudio(videoID)
		if err != nil {
			slog.Error("error configuring stream for track", "error", err.Error())
			return StreamReadyMsg{Data: data, Err: err}
		}
		slog.Info("stream configured sucessfully")
		return StreamReadyMsg{Data: data, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		fw := styles.ContainerFrameWidth()
		m.progress.Width = msg.Width - fw - 30
		if m.progress.Width < 10 {
			m.progress.Width = 10
		}

	case PlayerTickMsg:
		cmds = append(cmds, tickCmd())
		if m.status == playing || m.status == paused {
			pos, dur, pausd, finished, err := audio.GetProgress()
			if err == nil {
				m.position = pos
				if dur > 0 {
					m.duration = dur
				}

				if pausd {
					m.status = paused
				} else {
					m.status = playing
				}

				if m.duration > 0 {
					cmd := m.progress.SetPercent(pos / m.duration)
					cmds = append(cmds, cmd)

					if finished || (m.duration > 0 && pos >= m.duration-1) {
						cmds = append(cmds, func() tea.Msg { return NextTrackMsg{} })
					}
				}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case TogglePauseMsg:
		if m.status == playing || m.status == paused {
			audio.TogglePause()
			if m.status == playing {
				m.status = paused
			} else {
				m.status = playing
			}
		}

	case NextTrackMsg:
		if m.curTrack+1 < len(m.tracks) {
			cmds = append(cmds, m.playTrack(m.curTrack+1), m.spinner.Tick)
		}
	case PrevTrackMsg:
		if m.curTrack-1 >= 0 {
			cmds = append(cmds, m.playTrack(m.curTrack-1), m.spinner.Tick)
		}
	case QueueLoadedMsg:
		m.tracks = msg.Tracks
		m.continuation = msg.Continuation

		cmds = append(cmds, m.playTrack(msg.CurTrack), m.spinner.Tick)

	case StreamReadyMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.status = standby
			return m, nil
		}
		return m, func() tea.Msg {
			if err := audio.PlayStream(msg.Data); err != nil {
				return StreamErrorMsg{Err: err}
			}
			return StreamStartedMsg{}
		}

	case StreamStartedMsg:
		m.status = playing
		cmds = append(cmds, m.prefetchTrack(m.curTrack+1))

	case StreamErrorMsg:
		slog.Info("Error on streaming", "err", msg.Err)
		m.status = standby
		m.err = msg.Err

	case PrefetchCompleteMsg:
		if msg.Err == nil {
			if msg.Index < m.curTrack+2 {
				cmds = append(cmds, m.prefetchTrack(msg.Index+1))
			}
		}

	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	fw := styles.ContainerFrameWidth()
	w := m.width - fw
	sideWidth := w / 4
	centerWidth := w - (sideWidth * 2)

	// the player always has 2 lines
	var line1, line2 string
	if m.err != nil {
		line1 = lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(fmt.Sprintf("Error: %v", m.err))
	} else if m.status == standby || m.curTrack == -1 {
		line1 = lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render("▶ Not Playing - No track selected")
	} else if m.status == downloading {
		line1 = lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(m.spinner.View())
	} else {

		styleLeft := lipgloss.NewStyle().Width(sideWidth).Align(lipgloss.Left)
		styleCenter := lipgloss.NewStyle().Width(centerWidth).Align(lipgloss.Center)
		styleRight := lipgloss.NewStyle().Width(sideWidth).Align(lipgloss.Right)

		tr := m.tracks[m.curTrack]
		artistName := ""
		if len(tr.Artists) > 0 {
			artistName = tr.Artists[0].Name
		}

		statusIcon := "󰒮 󰏤 󰒭"
		if m.status == playing {
			statusIcon = "󰒮 󰐊 󰒭"
		}

		maxBar := 40
		barWidth := centerWidth - 22 // Espaço para "00:00 [bar] 00:00"
		if barWidth > maxBar {
			barWidth = maxBar
		}
		if barWidth < 10 {
			barWidth = 10
		}
		m.progress.Width = barWidth

		songName := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render(styles.Truncate(tr.Name, sideWidth-2))
		artist := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(styles.Truncate(artistName, sideWidth-2))
		icon := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(statusIcon)
		volumeStr := fmt.Sprintf("%s %d%%", m.getVolumeIcon(), audio.VolumePercent())

		line1 = lipgloss.JoinHorizontal(lipgloss.Center,
			styleLeft.Render(songName),
			styleCenter.Render(icon),
			styleRight.Render(volumeStr),
		)

		progressWidget := lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().MarginRight(1).Render(formatDuration(m.position)),
			m.progress.View(),
			lipgloss.NewStyle().MarginLeft(1).Render(formatDuration(m.duration)),
		)

		line2 = lipgloss.JoinHorizontal(lipgloss.Top,
			styleLeft.Render(artist),
			styleCenter.Render(progressWidget),
			styleRight.Render(""),
		)

	}

	return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
}

func (m Model) Height() int {
	return lipgloss.Height(m.View()) + styles.ContainerFrameHeight()
}

// getVolumeIcon returns the volume icon based on current audio.VolumePercent()
func (m Model) getVolumeIcon() string {
	v := audio.VolumePercent()
	if v == 0 {
		return ""
	}
	if v > 50 {
		return ""
	}
	return ""
}
