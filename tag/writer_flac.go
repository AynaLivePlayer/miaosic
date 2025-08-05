package tag

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"os"
)

type posMetaBlock[T any] struct {
	block T
	idx   int
}

func WriteFlacTags(f *os.File, meta Metadata) error {
	flacFile, err := flac.ParseBytes(f)
	if err != nil {
		return fmt.Errorf("error parsing flac file: %w", err)
	}
	var commentBlock posMetaBlock[*flacvorbis.MetaDataBlockVorbisComment]
	var pictures = map[byte]posMetaBlock[*flacpicture.MetadataBlockPicture]{}
	var pic *flacpicture.MetadataBlockPicture
	var cmt *flacvorbis.MetaDataBlockVorbisComment
	for idx, metaBlock := range flacFile.Meta {
		if metaBlock.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*metaBlock)
			if err == nil {
				commentBlock = posMetaBlock[*flacvorbis.MetaDataBlockVorbisComment]{
					block: cmt,
					idx:   idx,
				}
			}
		}
		if metaBlock.Type == flac.Picture {
			pic, err = flacpicture.ParseFromMetaDataBlock(*metaBlock)
			if err == nil {
				pictures[byte(pic.PictureType)] = posMetaBlock[*flacpicture.MetadataBlockPicture]{
					block: pic,
					idx:   idx,
				}
			}
		}
	}
	// write comment, include basic info and lyrcis
	commentBlockExists := true
	if commentBlock.block == nil {
		commentBlock.block = &flacvorbis.MetaDataBlockVorbisComment{
			Comments: []string{},
		}
		commentBlockExists = false
	}
	// just reset all
	commentBlock.block.Vendor = "miaosic" + miaosic.VERSION
	commentBlock.block.Comments = []string{}
	_ = commentBlock.block.Add(flacvorbis.FIELD_TITLE, meta.Title)
	_ = commentBlock.block.Add(flacvorbis.FIELD_ARTIST, meta.Artist)
	_ = commentBlock.block.Add(flacvorbis.FIELD_ALBUM, meta.Album)
	for _, lyric := range meta.Lyrics {
		_ = commentBlock.block.Add("LYRICS", lyric.Lyrics)
	}
	commentBlockMeta := commentBlock.block.Marshal()
	if commentBlockExists {
		flacFile.Meta[commentBlock.idx] = &commentBlockMeta
	} else {
		flacFile.Meta = append(flacFile.Meta, &commentBlockMeta)
	}
	// write file
	for _, picture := range meta.Pictures {
		newPic, err := flacpicture.NewFromImageData(flacpicture.PictureType(picture.Type),
			picture.Description, picture.Data, picture.Mimetype)
		if err != nil {
			continue
		}
		picBlock, ok := pictures[picture.Type]
		picBlockMeta := newPic.Marshal()
		if ok {
			flacFile.Meta[picBlock.idx] = &picBlockMeta
		} else {
			flacFile.Meta = append(flacFile.Meta, &picBlockMeta)
		}
	}
	return flacFile.Save(f.Name())
}
