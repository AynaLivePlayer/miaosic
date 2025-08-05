package cmds

import (
	"bytes"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/cmd/miaosic/cmds/tagwriter"
	"github.com/gabriel-vasile/mimetype"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	writeMetadata   bool
	downloadQuality string
)

func init() {
	CmdDownload.Flags().BoolVar(&writeMetadata, "metadata", true, "Write metadata (tags, cover, lyrics) to the file")
	CmdDownload.Flags().StringVar(&downloadQuality, "quality", "", "Quality preference (e.g., 128k, 320k, flac)")
}

var CmdDownload = &cobra.Command{
	Use:   "download <provider> <uri>",
	Short: "Download media, with metadata and cover art",
	Long: `Downloads a media file from a provider.
It fetches media information, URL, lyrics, and cover art.
By default, it writes all available metadata to the downloaded file.
Supported formats for metadata include MP3 and FLAC.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Steps 1-3: Get provider, media info, and URL (this part is unchanged)
		providerName := args[0]
		uri := args[1]
		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			return
		}
		meta, ok := provider.MatchMedia(uri)
		if !ok {
			fmt.Printf("URI not matched by provider: %s\n", uri)
			return
		}
		fmt.Println("Fetching media info...")
		info, err := provider.GetMediaInfo(meta)
		if err != nil {
			fmt.Printf("Error getting media info: %v\n", err)
			return
		}
		fmt.Printf("Found: %s - %s\n", info.Artist, info.Title)

		fmt.Println("Fetching media URL...")
		urls, err := provider.GetMediaUrl(meta, miaosic.Quality(downloadQuality))
		if err != nil || len(urls) == 0 {
			fmt.Printf("Error getting media URL or no URL found: %v\n", err)
			return
		}
		mediaURL := urls[0]
		fmt.Printf("Selected quality: %s\n", mediaURL.Quality)

		// Step 4: Download media file with progress bar (this part is unchanged)
		resp, err := http.Get(mediaURL.Url) // Simplified GET for clarity
		if err != nil {
			fmt.Printf("Error starting download: %v\n", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error downloading file: server responded with status %d\n", resp.StatusCode)
			return
		}
		totalSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		bar := progressbar.NewOptions64(totalSize,
			progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s...", info.Title)),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		)

		// Create a buffer and copy the download into it, with progress
		mediaData := &bytes.Buffer{}
		pRead := progressbar.NewReader(resp.Body, bar)
		_, err = io.Copy(mediaData, &pRead)
		if err != nil {
			fmt.Printf("Error during download: %v\n", err)
			return
		}

		// *** NEW: Detect content type from the first 512 bytes ***
		downloadedBytes := mediaData.Bytes()
		detectedContentType := mimetype.Detect(downloadedBytes[:min(512, len(downloadedBytes))]).String()

		ext, err := extensionFromContentType(detectedContentType)
		if err != nil {
			// Fallback strategy if detection is inconclusive
			fmt.Printf("Warning: Could not determine file type from content (%s). Falling back to URL extension.\n", detectedContentType)
			parsedURL, urlErr := url.Parse(mediaURL.Url)
			if urlErr != nil {
				// If URL is malformed, we can't get an extension.
				ext = ""
			} else {
				// Get extension from the path, which has no query string.
				ext = filepath.Ext(parsedURL.Path)
			}
			if ext == "" {
				fmt.Println("Warning: Could not determine file type from URL. Defaulting to .mp3.")
				ext = ".mp3" // Final fallback
			}
		}
		fmt.Printf("Detected file type: %s (%s)\n", detectedContentType, ext)

		// Step 5: Save the file from the buffer
		filename := sanitizeFilename(fmt.Sprintf("%s - %s%s", info.Artist, info.Title, ext))
		err = os.WriteFile(filename, downloadedBytes, 0644)
		if err != nil {
			fmt.Printf("Error saving file to disk: %v\n", err)
			return
		}

		// If metadata writing is disabled, we are done.
		if !writeMetadata {
			fmt.Printf("Download complete! Saved to %s\n", filename)
			return
		}

		// Step 6: Write Metadata (same as before)
		fmt.Println("Writing metadata...")
		lyric, _ := provider.GetMediaLyric(meta)
		var coverData []byte
		if info.Cover.Url != "" {
			fmt.Println("Downloading cover art...")
			coverResp, err := http.Get(info.Cover.Url)
			if err == nil && coverResp.StatusCode == http.StatusOK {
				defer coverResp.Body.Close()
				coverData, _ = io.ReadAll(coverResp.Body)
				info.Cover.Data = coverData
				fmt.Println("Cover art downloaded.")
			} else {
				fmt.Println("Could not download cover art.")
			}
		}
		err = tagFile(filename, ext, info, lyric, info.Cover)
		if err != nil {
			fmt.Printf("Error writing metadata: %v\n", err)
			fmt.Println("File is saved without metadata.")
		} else {
			fmt.Println("Metadata written successfully.")
		}
		fmt.Printf("Download complete! Saved to %s\n", filename)
	},
}

// tagFile and its helpers (tagMp3, tagFlac) remain unchanged.
func tagFile(filename, ext string, info miaosic.MediaInfo, lyric []miaosic.Lyrics, cover miaosic.Picture) error {
	switch strings.ToLower(ext) {
	case ".mp3":
		return tagwriter.WriteId3v2(filename, info, lyric, cover)
	case ".flac":
		return tagwriter.WriteFlac(filename, info, lyric, cover)
	default:
		return fmt.Errorf("unsupported file type for tagging: %s", ext)
	}
}

func extensionFromContentType(ct string) (string, error) {
	switch ct {
	case "audio/mpeg":
		return ".mp3", nil
	case "audio/flac", "audio/x-flac":
		return ".flac", nil
	case "audio/mp4":
		return ".m4a", nil
	case "audio/aac":
		return ".aac", nil
	}
	return "", fmt.Errorf("unsupported content type: %s", ct)
}
