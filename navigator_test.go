package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary directory structure for testing
func createTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "nav_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create some test files and directories
	os.MkdirAll(filepath.Join(tempDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".hidden_file"), []byte("content"), 0644)

	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

func TestNewNavigator(t *testing.T) {
	// Test with a valid path
	nav, err := NewNavigator(".")
	if err != nil {
		t.Errorf("NewNavigator failed for valid path: %v", err)
	}
	if nav.GetCurrentPath() == "" {
		t.Error("NewNavigator currentPath is empty for valid path")
	}

	// Test with a non-existent path
	nav, err = NewNavigator("/non/existent/path/123")
	if err != nil {
		t.Errorf("NewNavigator returned error for non-existent path: %v", err)
	}

	// ScanDirectory should fail for a non-existent path
	err = nav.ScanDirectory()
	if err == nil {
		t.Error("ScanDirectory did not return error for non-existent path")
	}
}

func TestScanDirectory(t *testing.T) {
	tempDir, cleanup := createTestDir(t)
	defer cleanup()

	nav, err := NewNavigator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create navigator for test dir: %v", err)
	}

	err = nav.ScanDirectory()
	if err != nil {
		t.Errorf("ScanDirectory failed: %v", err)
	}

	items := nav.GetItems()
	if len(items) == 0 {
		t.Error("ScanDirectory returned no items")
	}

	// Collect names from scanned items
	scannedNames := make([]string, len(items))
	for i, item := range items {
		scannedNames[i] = item.Name
	}

	// Assert presence of key items
	assertContains(t, scannedNames, "../")
	assertContains(t, scannedNames, "dir1")
	assertContains(t, scannedNames, "dir2")
	assertContains(t, scannedNames, "file1.txt")
	assertContains(t, scannedNames, ".hidden_file")

	// Verify item counts
	dirCount := 0
	fileCount := 0
	hiddenFileCount := 0
	for _, item := range items {
		if item.IsDir && item.Name != "../" {
			dirCount++
		}
		if !item.IsDir && !item.IsHidden {
			fileCount++
		}
		if item.IsHidden {
			hiddenFileCount++
		}
	}

	if dirCount != 2 {
		t.Errorf("Expected 2 directories, got %d", dirCount)
	}
	if fileCount != 1 {
		t.Errorf("Expected 1 file, got %d", fileCount)
	}
	if hiddenFileCount != 1 {
		t.Errorf("Expected 1 hidden file, got %d", hiddenFileCount)
	}
}

func TestMoveSelection(t *testing.T) {
	tempDir, cleanup := createTestDir(t)
	defer cleanup()

	nav, _ := NewNavigator(tempDir)
	nav.ScanDirectory()

	initialIdx := nav.GetSelectedIndex()

	// Move down
	nav.MoveSelection(1)
	if nav.GetSelectedIndex() != initialIdx+1 {
		t.Errorf("MoveSelection(1) failed, expected %d, got %d", initialIdx+1, nav.GetSelectedIndex())
	}

	// Move up
	nav.MoveSelection(-1)
	if nav.GetSelectedIndex() != initialIdx {
		t.Errorf("MoveSelection(-1) failed, expected %d, got %d", initialIdx, nav.GetSelectedIndex())
	}

	// Test boundary conditions
	items := nav.GetItems()
	nav.MoveSelection(-10) // Try to move way up
	if nav.GetSelectedIndex() != 0 {
		t.Errorf("MoveSelection beyond bounds failed, expected 0, got %d", nav.GetSelectedIndex())
	}

	nav.MoveSelection(len(items) + 10) // Try to move way down
	if nav.GetSelectedIndex() != len(items)-1 {
		t.Errorf("MoveSelection beyond bounds failed, expected %d, got %d", len(items)-1, nav.GetSelectedIndex())
	}
}

func TestSearchFunctionality(t *testing.T) {
	tempDir, cleanup := createTestDir(t)
	defer cleanup()

	nav, _ := NewNavigator(tempDir)
	nav.ScanDirectory()

	// Test initial state (not in search mode)
	if nav.GetSearchMode() {
		t.Error("Initially, searchMode should be false")
	}

	// Toggle search mode on
	nav.ToggleSearchMode()
	if !nav.GetSearchMode() {
		t.Error("ToggleSearchMode failed to set searchMode to true")
	}

	// Set search term and check filtering
	nav.SetSearchTerm("file")
	filteredItems := nav.GetItems()
	expectedFilteredNames := []string{".hidden_file", "file1.txt"}
	assertContainsAll(t, filteredItems, expectedFilteredNames)

	// Test case-insensitivity
	nav.SetSearchTerm("FILE")
	filteredItems = nav.GetItems()
	assertContainsAll(t, filteredItems, expectedFilteredNames)

	// Test no match
	nav.SetSearchTerm("nomatch")
	filteredItems = nav.GetItems()
	if len(filteredItems) != 0 {
		t.Errorf("Filtering for 'nomatch' failed, expected empty, got %v", filteredItems)
	}

	// Toggle search mode off
	nav.ToggleSearchMode()
	if nav.GetSearchMode() {
		t.Error("ToggleSearchMode failed to set searchMode to false")
	}
	if nav.GetSearchTerm() != "" {
		t.Errorf("searchTerm not cleared after exiting search mode, got %q", nav.GetSearchTerm())
	}
}

func TestGetSelectedItem(t *testing.T) {
	tempDir, cleanup := createTestDir(t)
	defer cleanup()

	nav, _ := NewNavigator(tempDir)
	nav.ScanDirectory()

	// Test getting selected item
	selectedItem := nav.GetSelectedItem()
	if selectedItem == nil {
		t.Error("GetSelectedItem returned nil")
	}

	// Test with empty items (edge case)
	nav.SetSearchTerm("nomatch") // This should result in empty filtered items
	selectedItem = nav.GetSelectedItem()
	if selectedItem != nil {
		t.Error("GetSelectedItem should return nil for empty items")
	}
}

// Helper functions
func assertContains(t *testing.T, slice []string, item string) {
	found := false
	for _, s := range slice {
		if s == item {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected slice to contain %q, but it did not. Slice: %v", item, slice)
	}
}

func assertContainsAll(t *testing.T, items []FileItem, expectedNames []string) {
	foundCount := 0
	for _, expectedName := range expectedNames {
		found := false
		for _, item := range items {
			if item.Name == expectedName {
				found = true
				foundCount++
				break
			}
		}
		if !found {
			t.Errorf("Expected item %q not found in filtered items", expectedName)
		}
	}
	if foundCount != len(expectedNames) {
		t.Errorf("Expected %d items, but found %d", len(expectedNames), foundCount)
	}
}