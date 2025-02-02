package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

// Directory represents a node_modules directory with its size
type Directory struct {
	path string
	size int64
}

// calculateDirSize calculates the total size of a directory
func calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// findNodeModules finds all node_modules directories concurrently
func findNodeModules(root string) ([]Directory, error) {
	var (
		nodeModules []Directory
		mutex       sync.Mutex
		wg          sync.WaitGroup
	)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue walking
		}

		if info.IsDir() && info.Name() == "node_modules" {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				size, err := calculateDirSize(p)
				if err == nil {
					mutex.Lock()
					nodeModules = append(nodeModules, Directory{path: p, size: size})
					mutex.Unlock()
				}
			}(path)
			return filepath.SkipDir
		}
		return nil
	})

	wg.Wait()
	return nodeModules, err
}

// formatSize converts bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// deleteDirectory deletes a directory with progress feedback
func deleteDirectory(dir Directory) error {
	start := time.Now()
	err := os.RemoveAll(dir.path)
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", dir.path, err)
	}

	fmt.Printf("Deleted [%s] (%s) in %s ✅\n",
		dir.path,
		formatSize(dir.size),
		duration.Round(time.Millisecond))
	return nil
}

func main() {
	var root string
	if len(os.Args) > 1 {
		root = os.Args[1]
	} else {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
	}
	fmt.Printf("Scanning for node_modules in %s (this may take a moment)...\n", root)

	// Find all node_modules directories with their sizes
	dirs, err := findNodeModules(root)
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	if len(dirs) == 0 {
		fmt.Printf("No node_modules directories found in %s\n", root)
		return
	}

	// Create options with sizes
	var options []string
	for _, dir := range dirs {
		options = append(options, fmt.Sprintf("%s (%s)", dir.path, formatSize(dir.size)))
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message:  fmt.Sprintf("Found %d node_modules directories. Select directories to DELETE:", len(dirs)),
		Options:  options,
		PageSize: 50,
	}

	if err = survey.AskOne(prompt, &selectedIndices); err != nil {
		fmt.Printf("Error during selection: %v\n", err)
		return
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No directories selected for deletion.")
		return
	}

	// Calculate total size to be deleted
	var totalSize int64
	for _, idx := range selectedIndices {
		totalSize += dirs[idx].size
	}

	// Confirm deletion with total size
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to DELETE %d directories (total size: %s)? This cannot be undone!",
			len(selectedIndices),
			formatSize(totalSize)),
	}

	if err = survey.AskOne(confirmPrompt, &confirm); err != nil {
		fmt.Printf("Error during confirmation: %v\n", err)
		return
	}

	if !confirm {
		fmt.Println("Operation cancelled.")
		return
	}

	fmt.Printf("\nDeleting %d directories (total size: %s) ⏳\n", len(selectedIndices), formatSize(totalSize))

	// Delete directories concurrently with a worker pool
	const maxConcurrent = 3
	semaphore := make(chan struct{}, maxConcurrent)
	var deleteWg sync.WaitGroup

	for _, idx := range selectedIndices {
		deleteWg.Add(1)
		go func(dir Directory) {
			defer deleteWg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := deleteDirectory(dir); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
		}(dirs[idx])
	}

	deleteWg.Wait()
	fmt.Println("\nOperation completed! 🎉")
}
