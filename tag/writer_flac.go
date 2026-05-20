package tag

import (
	"fmt"
	"os"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
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
	var cmt *flacvorbis.MetaDataBlockVorbisComment
	metaBlocks := make([]*flac.MetaDataBlock, 0, len(flacFile.Meta))
	for _, metaBlock := range flacFile.Meta {
		if metaBlock.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*metaBlock)
			if err == nil {
				commentBlock = posMetaBlock[*flacvorbis.MetaDataBlockVorbisComment]{
					block: cmt,
					idx:   len(metaBlocks),
				}
			}
		}
		if metaBlock.Type == flac.Picture {
			continue
		}
		metaBlocks = append(metaBlocks, metaBlock)
	}
	flacFile.Meta = metaBlocks
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
		picBlockMeta := newPic.Marshal()
		flacFile.Meta = append(flacFile.Meta, &picBlockMeta)
	}
	return flacFile.Save(f.Name())
}
