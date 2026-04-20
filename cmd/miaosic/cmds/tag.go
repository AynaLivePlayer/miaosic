package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AynaLivePlayer/miaosic/tag"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

var (
	tagReadFormat string

	tagWriteTitle    string
	tagWriteArtist   string
	tagWriteAlbum    string
	tagWriteLyrics   string
	tagWriteLang     string
	tagWriteCover    string
	tagWriteCoverTyp byte
)

func init() {
	CmdTagRead.Flags().StringVar(&tagReadFormat, "format", "plain", "output format: plain or json")

	CmdTagWrite.Flags().StringVar(&tagWriteTitle, "title", "", "title")
	CmdTagWrite.Flags().StringVar(&tagWriteArtist, "artist", "", "artist")
	CmdTagWrite.Flags().StringVar(&tagWriteAlbum, "album", "", "album")
	CmdTagWrite.Flags().StringVar(&tagWriteLyrics, "lyrics", "", "lyrics text")
	CmdTagWrite.Flags().StringVar(&tagWriteLang, "lyrics-lang", "eng", "lyrics language")
	CmdTagWrite.Flags().StringVar(&tagWriteCover, "cover", "", "cover image path")
	CmdTagWrite.Flags().Uint8Var(&tagWriteCoverTyp, "cover-type", tag.PictureTypeFrontCover, "cover picture type")

	CmdTag.AddCommand(CmdTagRead)
	CmdTag.AddCommand(CmdTagWrite)
}

var CmdTag = &cobra.Command{
	Use:   "tag",
	Short: "Read or write audio metadata tags",
	Long:  "Read or update tags (title, artist, album, lyrics, cover) for local audio files.",
}

var CmdTagRead = &cobra.Command{
	Use:     "read <file>",
	Short:   "Read metadata tags from audio file",
	Long:    "Read metadata tags from a local audio file and print in plain or JSON format.",
	Example: "  miaosic tag read ./data/test.wav\n  miaosic tag read ./data/test.mp3 --format json",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filepath := args[0]
		f, err := os.Open(filepath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer f.Close()

		meta, err := tag.Read(f)
		if err != nil {
			fmt.Printf("Error reading tags: %v\n", err)
			return
		}

		switch strings.ToLower(tagReadFormat) {
		case "json":
			b, err := json.MarshalIndent(meta, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling metadata: %v\n", err)
				return
			}
			fmt.Println(string(b))
		case "plain", "":
			printTagMetadata(meta)
		default:
			fmt.Printf("Unsupported format: %s (supported: plain, json)\n", tagReadFormat)
		}
	},
}

var CmdTagWrite = &cobra.Command{
	Use:   "write <file>",
	Short: "Write metadata tags to audio file",
	Long: `Write selected metadata fields to a local audio file.
Existing tags are preserved unless the corresponding flag is provided.`,
	Example: "  miaosic tag write ./song.mp3 --title \"Hello\" --artist \"A\"\n  miaosic tag write ./song.m4a --lyrics \"line1\" --lyrics-lang eng\n  miaosic tag write ./song.flac --cover ./data/cover.jpg --cover-type 3",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filepath := args[0]
		if tagWriteTitle == "" && tagWriteArtist == "" && tagWriteAlbum == "" && tagWriteLyrics == "" && tagWriteCover == "" {
			fmt.Println("Nothing to write. Use at least one of --title/--artist/--album/--lyrics/--cover.")
			return
		}

		meta := tag.Metadata{}
		if f, err := os.Open(filepath); err == nil {
			if existing, readErr := tag.Read(f); readErr == nil {
				meta = existing
			}
			_ = f.Close()
		}

		if tagWriteTitle != "" {
			meta.Title = tagWriteTitle
		}
		if tagWriteArtist != "" {
			meta.Artist = tagWriteArtist
		}
		if tagWriteAlbum != "" {
			meta.Album = tagWriteAlbum
		}
		if tagWriteLyrics != "" {
			meta.Lyrics = []tag.Lyrics{{
				Lang:   tagWriteLang,
				Lyrics: tagWriteLyrics,
			}}
		}
		if tagWriteCover != "" {
			coverData, err := os.ReadFile(tagWriteCover)
			if err != nil {
				fmt.Printf("Error reading cover image: %v\n", err)
				return
			}
			meta.Pictures = []tag.Picture{{
				Mimetype:    mimetype.Detect(coverData).String(),
				Type:        tagWriteCoverTyp,
				Description: "cover",
				Data:        coverData,
			}}
		}

		if err := tag.WriteTo(filepath, meta); err != nil {
			fmt.Printf("Error writing tags: %v\n", err)
			return
		}
		fmt.Println("Tags written successfully.")
	},
}

func printTagMetadata(meta tag.Metadata) {
	fmt.Println("Format:", meta.Format)
	fmt.Println("Mimetype:", meta.Mimetype)
	fmt.Println("Title:", meta.Title)
	fmt.Println("Artist:", meta.Artist)
	fmt.Println("Album:", meta.Album)
	if len(meta.Lyrics) == 0 {
		fmt.Println("Lyrics: <none>")
	} else {
		for i, lyric := range meta.Lyrics {
			fmt.Printf("Lyrics[%d]: [%s] %s\n", i, lyric.Lang, lyric.Lyrics)
		}
	}
	if len(meta.Pictures) == 0 {
		fmt.Println("Pictures: <none>")
	} else {
		for i, pic := range meta.Pictures {
			fmt.Printf("Picture[%d]: type=%d mime=%s size=%d\n", i, pic.Type, pic.Mimetype, len(pic.Data))
		}
	}
}
