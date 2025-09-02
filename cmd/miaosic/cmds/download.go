package cmds

import (
	"bytes"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/tag"
	"github.com/gabriel-vasile/mimetype"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

var (
	writeMetadata     bool
	downloadQuality   string
	specifiedFilename string
)

func init() {
	CmdDownload.Flags().BoolVar(&writeMetadata, "metadata", false, "Write metadata (tags, cover, lyrics) to the file")
	CmdDownload.Flags().StringVar(&downloadQuality, "quality", "", "Quality preference (e.g., 128k, 320k, flac)")
	CmdDownload.Flags().StringVar(&specifiedFilename, "filename", "", "Filename to use for download")
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
		resp, err := http.Get(mediaURL.Url)
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

		downloadedBytes := mediaData.Bytes()
		parsedURL, urlErr := url.Parse(mediaURL.Url)
		var ext string
		if urlErr != nil {
			ext = ""
		} else {
			ext = filepath.Ext(parsedURL.Path)
		}
		if ext == "" {
			ext = mimetype.Detect(downloadedBytes[:min(512, len(downloadedBytes))]).Extension()
		}
		filename := sanitizeFilename(fmt.Sprintf("%s - %s%s", info.Artist, info.Title, ext))
		// Step 5: Save the file from the buffer
		if specifiedFilename != "" {
			filename = specifiedFilename
		}

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
		tagMeta := tag.Metadata{
			Title:    info.Title,
			Artist:   info.Artist,
			Album:    info.Album,
			Lyrics:   nil,
			Pictures: nil,
		}
		for _, lrc := range lyric {
			tagMeta.Lyrics = append(tagMeta.Lyrics, tag.Lyrics{
				Lang:   lrc.Lang,
				Lyrics: lrc.String(),
			})
		}
		if info.Cover.Data != nil {
			tagMeta.Pictures = append(tagMeta.Pictures, tag.Picture{
				Mimetype:    mimetype.Detect(info.Cover.Data).String(),
				Type:        tag.PictureTypeFrontCover,
				Description: "AlbumCover",
				Data:        info.Cover.Data,
			})
		}

		err = tag.WriteTo(filename, tagMeta)
		if err != nil {
			fmt.Printf("Error writing metadata: %v\n", err)
			fmt.Println("File is saved without metadata.")
		} else {
			fmt.Println("Metadata written successfully.")
		}
		fmt.Printf("Download complete! Saved to %s\n", filename)
	},
}
