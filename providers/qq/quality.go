package qq

import (
	"github.com/AynaLivePlayer/miaosic"
	"strings"
)

const (
	QualityMaster  = "AI00.flac" // MASTER: 臻品母带2.0,24Bit 192kHz,size_new[0]
	QualityAtmos2  = "Q000.flac" // ATMOS_2: 臻品全景声2.0,16Bit 44.1kHz,size_new[1]
	QualityAtmos51 = "Q001.flac" // ATMOS_51: 臻品音质2.0,16Bit 44.1kHz,size_new[2]
	QualityFLAC    = "F000.flac" // FLAC: flac 格式,16Bit 44.1kHz~24Bit 48kHz,size_flac
	QualityOGG640  = "O801.ogg"  // OGG_640: ogg 格式,640kbps,size_new[5]
	QualityOGG320  = "O800.ogg"  // OGG_320: ogg 格式,320kbps,size_new[3]
	QualityOGG192  = "O600.ogg"  // OGG_192: ogg 格式,192kbps,size_192ogg
	QualityOGG96   = "O400.ogg"  // OGG_96: ogg 格式,96kbps,size_96ogg
	QualityMP3320  = "M800.mp3"  // MP3_320: mp3 格式,320kbps,size_320mp3
	QualityMP3128  = "M500.mp3"  // MP3_128: mp3 格式,128kbps,size_128mp3
	QualityACC192  = "C600.m4a"  // ACC_192: m4a 格式,192kbps,size_192aac
	QualityACC96   = "C400.m4a"  // ACC_96: m4a 格式,96kbps,size_96aac
	QualityACC48   = "C200.m4a"  // ACC_48: m4a 格式,48kbps,size_48aac
)

const (
	QualityEncMaster  = "AIM0.mflac" // MASTER: 臻品母带2.0,24Bit 192kHz,size_new[0]
	QualityEncAtmos2  = "Q0M0.mflac" // ATMOS_2: 臻品全景声2.0,16Bit 44.1kHz,size_new[1]
	QualityEncAtmos51 = "Q0M1.mflac" // ATMOS_51: 臻品音质2.0,16Bit 44.1kHz,size_new[2]
	QualityEncFLAC    = "F0M0.mflac" // FLAC: mflac 格式,16Bit 44.1kHz~24Bit 48kHz,size_flac
	QualityEncOGG640  = "O801.mgg"   // OGG_640: mgg 格式,640kbps,size_new[5]
	QualityEncOGG320  = "O800.mgg"   // OGG_320: mgg 格式,320kbps,size_new[3]
	QualityEncOGG192  = "O6M0.mgg"   // OGG_192: mgg 格式,192kbps,size_192ogg
	QualityEncOGG96   = "O4M0.mgg"   // OGG_96: mgg 格式,96kbps,size_96ogg
)

func IsQqQuality(quality miaosic.Quality) bool {
	val := strings.Split(string(quality), ".")
	return len(val) == 2
}

func isEncryptedQuality(quality miaosic.Quality) bool {
	val := strings.Split(string(quality), ".")
	if len(val) != 2 {
		return false
	}
	if val[1] == "mflac" || val[1] == "mgg" {
		return true
	}
	return false
}
