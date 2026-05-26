package cmd

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixUploadCmd = &cobra.Command{
	Use:   "upload <room-id> <file-path>",
	Short: "Upload and send a file to a room",
	Example: `  anime matrix upload '!abc:localhost' ./report.pdf
  anime matrix upload '#general:localhost' ./screenshot.png`,
	Args: cobra.ExactArgs(2),
	RunE: runMatrixUpload,
}

func init() {
	matrixCmd.AddCommand(matrixUploadCmd)
}

func runMatrixUpload(cmd *cobra.Command, args []string) error {
	roomID := args[0]
	filePath := args[1]

	cfg, _ := matrixcfg.Load()
	client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)

	// Resolve alias
	if strings.HasPrefix(roomID, "#") {
		resolved, err := client.ResolveAlias(roomID)
		if err != nil {
			return err
		}
		roomID = resolved
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fmt.Printf("  %s %s %s (%d bytes)\n", theme.SymbolLoading, theme.InfoStyle.Render("Uploading"), theme.HighlightStyle.Render(filename), len(data))

	// Upload
	mxcURL, err := client.UploadMedia(data, filename, contentType)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Uploaded"), theme.DimTextStyle.Render(mxcURL))

	// Send as appropriate type
	isImage := strings.HasPrefix(contentType, "image/")
	var eventID string
	if isImage {
		eventID, err = client.SendImage(roomID, mxcURL, filename, contentType, int64(len(data)))
	} else {
		eventID, err = client.SendFile(roomID, mxcURL, filename, contentType, int64(len(data)))
	}
	if err != nil {
		return err
	}

	fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Sent to room"), theme.DimTextStyle.Render(eventID))
	return nil
}
