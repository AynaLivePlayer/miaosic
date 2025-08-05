package tag

import (
	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"io"
	"strings"
)

func ReadFLACTags(r io.ReadSeeker, mime string) (Metadata, error) {
	meta := Metadata{
		Mimetype: mime,
		Format:   FormatVORBIS,
	}
	metadata, err := flac.ParseMetadata(r)
	if err != nil {
		return Metadata{}, err
	}
	for _, block := range metadata.Meta {
		switch block.Type {
		case flac.VorbisComment:
			comment, err := flacvorbis.ParseFromMetaDataBlock(*block)
			if err != nil {
				continue
			}
			for _, tag := range comment.Comments {
				parts := strings.SplitN(tag, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.ToUpper(parts[0])
				value := parts[1]

				switch key {
				case flacvorbis.FIELD_TITLE:
					meta.Title = value
				case flacvorbis.FIELD_ARTIST:
					meta.Artist = value
				case flacvorbis.FIELD_ALBUM:
					meta.Album = value
				case "LYRICS":
					meta.Lyrics = append(meta.Lyrics, Lyrics{
						Lang:   "unk",
						Lyrics: value,
					})
				}
			}
		case flac.Picture:
			pic, err := flacpicture.ParseFromMetaDataBlock(*block)
			if err != nil {
				continue
			}
			meta.Pictures = append(meta.Pictures, Picture{
				Mimetype:    pic.MIME,
				Type:        byte(pic.PictureType),
				Description: pic.Description,
				Data:        pic.ImageData,
			})
		default:
			// do nothing
		}
	}
	return meta, nil
}
