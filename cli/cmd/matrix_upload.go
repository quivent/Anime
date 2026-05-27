package cmd

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
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

		t.Info(fmt.Sprintf("uploading %s  %s", t.Bold(t.Gold.S(filename)), t.Dim(fmt.Sprintf("(%d bytes)", len(data)))))

		fileID, err := client.UploadFile(channelID, data, filename, contentType)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		t.Ok("uploaded  " + t.Dim(fileID))

		post, err := client.CreatePost(channelID, "", []string{fileID})
		if err != nil {
			return fmt.Errorf("send failed: %w", err)
		}
		t.Ok("sent to channel  " + t.Dim(post.ID))
		return nil
	},
}

func init() {
	matrixCmd.AddCommand(matrixUploadCmd)
}
