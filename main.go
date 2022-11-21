package main

// An example Bubble Tea server. This will put an ssh session into alt screen
// and continually print up to date terminal information.

import (
	"context"
	"fmt"
	"log"
  "strings"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
  "github.com/charmbracelet/lipgloss"
  "github.com/charmbracelet/bubbles/viewport"
	"github.com/muesli/termenv"
)

const (
	host = "localhost"
	port = 23234
)

const useHighPerformanceRenderer = true

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			myCustomBubbleteaMiddleware(),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

// You can write your own custom bubbletea middleware that wraps tea.Program.
// Make sure you set the program input and output to ssh.Session.
func myCustomBubbleteaMiddleware() wish.Middleware {
	newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		go func() {
			for {
				<-time.After(1 * time.Second)
				p.Send(timeMsg(time.Now()))
			}
		}()
		return p
	}
	teaHandler := func(s wish.Session) *tea.Program {
    maincontent, err := os.ReadFile("heart.txt")
    if err != nil {
		    fmt.Println("could not load file:", err)
	      os.Exit(1)
	  }

		m := model{
	    state:       0,
      ready:       false,
			time:        time.Now(),
      maincontent: string(maincontent),
		}
		return newProg(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen())
	}
	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

var (
  subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
  heart  = "  ,ad8PPPP88b,     ,d88PPPP8ba,   \n d8P\"      \"Y8b, ,d8P\"      \"Y8b \ndP'           \"8a8\"           `Yd\n8(              \"              )8\nI8                             8I\n Yb,                         ,dP \n  \"8a,                     ,a8\"  \n    \"8a,                 ,a8\"    \n      \"Yba             adP\"      \n        `Y8a         a8P'        \n          `88,     ,88'          \n            \"8b   d8\"            \n             \"8b d8\"             \n              `888'              \n                \""

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).BorderForeground(lipgloss.Color("#7D56F4"))
	}()

  lineStyle = func() lipgloss.Style {
    return lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
  }()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()

  heartStyle = func() lipgloss.Style {
    return lipgloss.NewStyle().
			Padding(1, 3).
      Foreground(lipgloss.Color("#FF0000"))
  }()
)


// Just a generic tea.Model to demo terminal information of ssh.
type model struct {
	time   time.Time
  state  int  // 0 main

  //viewports
  ready  bool
  viewport viewport.Model

  // main
  maincontent  string
}

type timeMsg time.Time

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

  m.time = time.Now()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		} else if k := msg.String(); k == "tab" {
      m.state = m.state + 1 % 3
    } else if k := msg.String(); k == "shift-tab" {
      if m.state == 0 {
        m.state = 2
      } else {
        m.state = m.state - 1
      }
    }

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width - 1, msg.Height-verticalMarginHeight)
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
      dialog := lipgloss.Place(msg.Width - 3, max(21, msg.Height-verticalMarginHeight),
  			lipgloss.Center, lipgloss.Center,
  			heartStyle.Render(heart),
  			lipgloss.WithWhitespaceChars("猫咪"),
  			lipgloss.WithWhitespaceForeground(subtle),
  		)
      m.viewport.SetContent(dialog)

			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
      dialog := lipgloss.Place(msg.Width - 2, 21,
  			lipgloss.Center, lipgloss.Center,
  			heartStyle.Render(heart),
  			lipgloss.WithWhitespaceChars("猫咪"),
  			lipgloss.WithWhitespaceForeground(subtle),
  		)
      m.viewport.SetContent(dialog)
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) headerView() string {
	title := titleStyle.Render("Home")
  clock := titleStyle.Render(m.time.Format(time.RFC1123))
	line := lineStyle.Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)-lipgloss.Width(clock))))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, clock)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := lineStyle.Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
  // Header

  // Body

  //
  if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}
