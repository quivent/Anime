package cmd

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixUploadCmd = &cobra.Command{
	Use:   "upload <channel-id> <file-path>",
	Short: "Upload and send a file to a channel",
	Example: `  anime matrix upload <channel-id> ./report.pdf
  anime matrix upload <channel-id> ./screenshot.png`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelID := args[0]
		filePath := args[1]

		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

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

		fmt.Printf("  %s %s %s (%d bytes)\n",
			theme.SymbolLoading,
			theme.InfoStyle.Render("Uploading"),
			theme.HighlightStyle.Render(filename),
			len(data))

		fileID, err := client.UploadFile(channelID, data, filename, contentType)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		fmt.Printf("  %s %s %s\n",
			theme.SymbolSuccess,
			theme.SuccessStyle.Render("Uploaded"),
			theme.DimTextStyle.Render(fileID))

		// Send post with file
		post, err := client.CreatePost(channelID, "", []string{fileID})
		if err != nil {
			return fmt.Errorf("send failed: %w", err)
		}
		fmt.Printf("  %s %s %s\n",
			theme.SymbolSuccess,
			theme.SuccessStyle.Render("Sent to channel"),
			theme.DimTextStyle.Render(post.ID))
		return nil
	},
}

func init() {
	matrixCmd.AddCommand(matrixUploadCmd)
}
