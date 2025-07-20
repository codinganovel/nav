package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// FileItem represents a file or directory entry.
type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	IsHidden bool
}

// Navigator manages the state of the file navigator.
type Navigator struct {
	currentPath   string
	items         []FileItem
	filteredItems []FileItem
	selectedIdx   int
	searchMode    bool
	searchTerm    string
}

// NewNavigator creates a new Navigator instance.
func NewNavigator(startPath string) (*Navigator, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}
	return &Navigator{
		currentPath: absPath,
		selectedIdx: 0,
	}, nil
}

// ScanDirectory reads the contents of the current directory and populates the items slice.
func (n *Navigator) ScanDirectory() error {
	entries, err := os.ReadDir(n.currentPath)
	if err != nil {
		// Check if it's a permission error or other access issue
		if os.IsPermission(err) {
			return err // Will be handled by caller with user-friendly message
		}
		// Try to handle unrecognized root or other path issues
		if n.isRootPath(n.currentPath) {
			// If we can't read root, fallback to home directory
			homeDir, homeErr := os.UserHomeDir()
			if homeErr == nil {
				n.currentPath = homeDir
				return n.ScanDirectory()
			}
		}
		return err
	}

	n.items = []FileItem{}

	// Add parent directory if not at root
	if n.currentPath != "/" && n.currentPath != `C:\` {
		parentPath := filepath.Dir(n.currentPath)
		n.items = append(n.items, FileItem{
			Name:     "../",
			Path:     parentPath,
			IsDir:    true,
			IsHidden: false,
		})
	}

	// Add current directory entries
	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(n.currentPath, name)
		isDir := entry.IsDir()
		isHidden := len(name) > 0 && name[0] == '.'

		n.items = append(n.items, FileItem{
			Name:     name,
			Path:     fullPath,
			IsDir:    isDir,
			IsHidden: isHidden,
		})
	}

	// Sort items: directories first, then files, both alphabetically
	sort.Slice(n.items, func(i, j int) bool {
		itemI := n.items[i]
		itemJ := n.items[j]

		// Handle "../" always at the top
		if itemI.Name == "../" {
			return true
		}
		if itemJ.Name == "../" {
			return false
		}

		// Directories come before files
		if itemI.IsDir != itemJ.IsDir {
			return itemI.IsDir
		}

		// Alphabetical sort within category
		return itemI.Name < itemJ.Name
	})

	n.filterItems()
	return nil
}

// GetCurrentPath returns the current directory path.
func (n *Navigator) GetCurrentPath() string {
	return n.currentPath
}

// GetItems returns the filtered items for display.
func (n *Navigator) GetItems() []FileItem {
	return n.filteredItems
}

// GetSelectedIndex returns the current selected index.
func (n *Navigator) GetSelectedIndex() int {
	return n.selectedIdx
}

// GetSearchMode returns whether search mode is active.
func (n *Navigator) GetSearchMode() bool {
	return n.searchMode
}

// GetSearchTerm returns the current search term.
func (n *Navigator) GetSearchTerm() string {
	return n.searchTerm
}

// MoveSelection moves the selection index by delta.
func (n *Navigator) MoveSelection(delta int) {
	n.selectedIdx += delta
	if n.selectedIdx < 0 {
		n.selectedIdx = 0
	}
	if n.selectedIdx >= len(n.filteredItems) {
		n.selectedIdx = len(n.filteredItems) - 1
	}
}

// GetSelectedItem returns the currently selected item.
func (n *Navigator) GetSelectedItem() *FileItem {
	if len(n.filteredItems) == 0 || n.selectedIdx >= len(n.filteredItems) {
		return nil
	}
	return &n.filteredItems[n.selectedIdx]
}

// OpenSelected opens the selected item.
func (n *Navigator) OpenSelected() error {
	selectedItem := n.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	if selectedItem.IsDir {
		// Navigate into directory
		n.currentPath = selectedItem.Path
		n.selectedIdx = 0
		n.searchTerm = ""
		n.searchMode = false
		return n.ScanDirectory()
	} else {
		// Open file's parent directory in terminal
		return n.openInTerminal(selectedItem.Path, false)
	}
}

// OpenSelectedInTerminal opens the selected item in a new terminal.
func (n *Navigator) OpenSelectedInTerminal() error {
	selectedItem := n.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	return n.openInTerminal(selectedItem.Path, selectedItem.IsDir)
}

// ToggleSearchMode toggles search mode on/off.
func (n *Navigator) ToggleSearchMode() {
	n.searchMode = !n.searchMode
	if !n.searchMode {
		n.searchTerm = ""
		n.filterItems()
	}
}

// SetSearchTerm sets the search term and filters items.
func (n *Navigator) SetSearchTerm(term string) {
	n.searchTerm = term
	n.filterItems()
}

// filterItems filters items based on search term.
func (n *Navigator) filterItems() {
	if n.searchTerm == "" {
		n.filteredItems = n.items
	} else {
		n.filteredItems = []FileItem{}
		lowerSearchTerm := strings.ToLower(n.searchTerm)
		for _, item := range n.items {
			if strings.Contains(strings.ToLower(item.Name), lowerSearchTerm) {
				n.filteredItems = append(n.filteredItems, item)
			}
		}
	}

	// Reset selection if it's out of bounds
	if n.selectedIdx >= len(n.filteredItems) {
		n.selectedIdx = 0
	}
}

// detectTerminalCommand detects the appropriate terminal command to use.
func detectTerminalCommand() (string, []string) {
	// 1. Check $TERMINAL environment variable first (highest priority)
	if terminal := os.Getenv("TERMINAL"); terminal != "" {
		parts := strings.Fields(terminal)
		if len(parts) > 0 {
			return parts[0], parts[1:]
		}
	}

	// 2. Check $TERM_PROGRAM for known terminals
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		switch strings.ToLower(termProgram) {
		case "ghostty":
			return "ghostty", []string{}
		case "iterm.app":
			return "open", []string{"-a", "iTerm"}
		case "apple_terminal":
			return "open", []string{"-a", "Terminal"}
		case "wezterm":
			return "wezterm", []string{"start"}
		case "kitty":
			return "kitty", []string{}
		case "alacritty":
			return "alacritty", []string{}
		}
	}

	// 3. Fall back to OS-specific defaults
	switch runtime.GOOS {
	case "darwin": // macOS
		return "open", []string{"-a", "Terminal"}
	case "linux": // Linux
		return "gnome-terminal", []string{}
	case "windows": // Windows
		return "cmd", []string{"/c", "start", "cmd", "/k"}
	default:
		return "xterm", []string{}
	}
}

// openInTerminal opens a new terminal window at the given path.
func (n *Navigator) openInTerminal(path string, isDir bool) error {
	workingDir := path
	if !isDir {
		workingDir = filepath.Dir(path)
	}

	command, args := detectTerminalCommand()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if command == "open" {
			// Special handling for macOS 'open' command
			cmd = exec.Command(command, append(args, workingDir)...)
		} else {
			// For other terminals like ghostty, wezterm, etc.
			allArgs := append(args, "--working-directory", workingDir)
			cmd = exec.Command(command, allArgs...)
		}
	case "linux":
		if command == "gnome-terminal" {
			cmd = exec.Command(command, "--working-directory", workingDir)
		} else {
			// For other terminals, try common working directory flags
			allArgs := append(args, "--working-directory", workingDir)
			cmd = exec.Command(command, allArgs...)
		}
	case "windows":
		if command == "cmd" {
			// Special handling for Windows cmd
			allArgs := append(args, "cd", workingDir)
			cmd = exec.Command(command, allArgs...)
		} else {
			// For other terminals like Windows Terminal
			allArgs := append(args, "--starting-directory", workingDir)
			cmd = exec.Command(command, allArgs...)
		}
	default:
		// Generic Unix-like system
		allArgs := append(args, workingDir)
		cmd = exec.Command(command, allArgs...)
	}

	// Start the command in the background
	return cmd.Start()
}

// isRootPath checks if the given path is a root path that might cause issues
func (n *Navigator) isRootPath(path string) bool {
	// Check for common root paths that might not be accessible
	rootPaths := []string{"/", "C:\\", "D:\\", "E:\\", "F:\\"}
	for _, root := range rootPaths {
		if path == root {
			return true
		}
	}
	return false
}