package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func main() {
	// Handle help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		showHelp()
		return
	}

	// Get starting directory from command line or use current directory
	startPath := "."
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}

	// Initialize tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating screen: %v\n", err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing screen: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	// Set up default style
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	screen.SetStyle(defStyle)

	// Create navigator
	navigator, err := NewNavigator(startPath)
	if err != nil {
		screen.Fini()
		fmt.Fprintf(os.Stderr, "Error creating navigator: %v\n", err)
		os.Exit(1)
	}

	// Initial directory scan
	if err = navigator.ScanDirectory(); err != nil {
		screen.Fini()
		if os.IsPermission(err) {
			fmt.Fprintf(os.Stderr, "Permission denied: Cannot access directory '%s'\n", navigator.GetCurrentPath())
		} else {
			fmt.Fprintf(os.Stderr, "Cannot read directory '%s': %v\n", navigator.GetCurrentPath(), err)
		}
		os.Exit(1)
	}

	// Main event loop
	for {
		drawUI(screen, navigator, defStyle)

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if navigator.GetSearchMode() {
				if handleSearchModeKey(ev, navigator) {
					return // Exit requested
				}
			} else {
				if handleNormalModeKey(ev, navigator) {
					return // Exit requested
				}
			}
		case *tcell.EventResize:
			// Just redraw on resize
			continue
		}
	}
}

// handleSearchModeKey handles keyboard input in search mode.
func handleSearchModeKey(ev *tcell.EventKey, navigator *Navigator) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		navigator.ToggleSearchMode()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		searchTerm := navigator.GetSearchTerm()
		if len(searchTerm) > 0 {
			navigator.SetSearchTerm(searchTerm[:len(searchTerm)-1])
		}
	case tcell.KeyRune:
		searchTerm := navigator.GetSearchTerm()
		navigator.SetSearchTerm(searchTerm + string(ev.Rune()))
	}
	return false
}

// handleNormalModeKey handles keyboard input in normal mode.
func handleNormalModeKey(ev *tcell.EventKey, navigator *Navigator) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		navigator.MoveSelection(-1)
	case tcell.KeyDown:
		navigator.MoveSelection(1)
	case tcell.KeyEnter:
		if err := navigator.OpenSelected(); err != nil {
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "\nPermission denied: Cannot access the selected item\n")
			} else {
				fmt.Fprintf(os.Stderr, "\nError opening selected item: %v\n", err)
			}
		}
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			return true // Exit
		case '/':
			navigator.ToggleSearchMode()
		case 'o':
			if err := navigator.OpenSelectedInTerminal(); err != nil {
				fmt.Fprintf(os.Stderr, "\nError opening terminal: %v\n", err)
			}
		}
	}
	return false
}

// drawUI renders the current state to the screen.
func drawUI(screen tcell.Screen, navigator *Navigator, defStyle tcell.Style) {
	screen.Clear()
	_, h := screen.Size()

	// Draw current path
	drawText(screen, 0, 0, defStyle, navigator.GetCurrentPath())

	// Draw items
	items := navigator.GetItems()
	for i, item := range items {
		y := i + 2 // Start drawing items from y=2
		if y >= h-2 { // Leave space for status bar
			break
		}

		style := defStyle
		if i == navigator.GetSelectedIndex() {
			style = defStyle.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack)
		}

		// Draw tree-style prefix
		prefix := "├── "
		if i == len(items)-1 {
			prefix = "└── "
		}

		// Format display name
		displayName := item.Name
		if item.IsDir && displayName != "../" {
			displayName += "/"
		}

		drawText(screen, 0, y, style, prefix+displayName)
	}

	// Draw status bar
	statusBarY := h - 1
	statusContent := buildStatusBar(navigator, len(items))
	drawText(screen, 0, statusBarY, defStyle, statusContent)

	screen.Show()
}

// buildStatusBar builds the status bar content.
func buildStatusBar(navigator *Navigator, totalItems int) string {
	if navigator.GetSearchMode() {
		return fmt.Sprintf("Search: %s", navigator.GetSearchTerm())
	}
	return fmt.Sprintf("[%d items] • ↑↓ navigate • Enter open • o open in terminal • q quit • / search", totalItems)
}

// drawText draws text at the specified position.
func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	w, _ := screen.Size()
	
	// Smart truncation for long text
	if len(text) > w-x {
		text = truncateFilename(text, w-x-1)
	}
	
	for i, r := range []rune(text) {
		if x+i >= w {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}
}

// truncateFilename intelligently truncates long filenames
func truncateFilename(filename string, maxLen int) string {
	if len(filename) <= maxLen {
		return filename
	}
	
	// If it's too short to truncate meaningfully, just use ellipsis
	if maxLen < 10 {
		return filename[:maxLen-1] + "…"
	}
	
	// For filenames with extensions, try to preserve the extension
	if strings.Contains(filename, ".") && !strings.HasPrefix(filename, ".") {
		parts := strings.Split(filename, ".")
		if len(parts) >= 2 {
			ext := "." + parts[len(parts)-1]
			nameWithoutExt := strings.Join(parts[:len(parts)-1], ".")
			
			// If extension is reasonable length, preserve it
			if len(ext) <= maxLen/3 {
				availableForName := maxLen - len(ext) - 1 // -1 for ellipsis
				if availableForName > 0 {
					return nameWithoutExt[:availableForName] + "…" + ext
				}
			}
		}
	}
	
	// Default truncation
	return filename[:maxLen-1] + "…"
}

// showHelp displays help information.
func showHelp() {
	fmt.Print(`nav - Terminal File Navigator

USAGE:
  nav [directory]     Navigate to directory (default: current directory)
  nav --help, -h      Show this help

KEYBINDINGS:
  ↑/↓        Navigate up/down
  Enter      Open directory / Open file's parent in terminal
  o          Open selected item in new terminal
  /          Search (type to filter, Esc to exit)
  q          Quit

TERMINAL DETECTION:
  nav automatically detects your terminal:
  1. $TERMINAL environment variable (highest priority)
  2. $TERM_PROGRAM detection (iTerm2, Ghostty, Wezterm, etc.)
  3. OS defaults (Terminal.app, gnome-terminal, cmd)

  Examples:
    export TERMINAL="open -a Ghostty"
    export TERMINAL="wezterm start --cwd"
    export TERMINAL="alacritty --working-directory"

FEATURES:
  • Smart terminal detection
  • Real-time search filtering
  • Cross-platform support (macOS, Linux, Windows)
  • Tree-style directory display
  • Hidden file support

`)
}