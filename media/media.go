package media

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/leberKleber/go-mpris"
	"github.com/ztrue/tracerr"

	"xyosc/utils"
)

func FmtDuration(s float64) string {
	mf := int(math.Floor(s / 60))
	sf := int((s/60 - math.Floor(s/60)) * 60)
	return fmt.Sprintf("%02d:%02d", mf, sf)
}

type CurrentPlayingMediaInfo struct {
	Title    string
	Album    string
	Artist   string
	Position float64
	Duration float64
	Playing  bool
}

var PlayingMediaInfo CurrentPlayingMediaInfo

var lastInterpolationCallTime time.Time

func Interpolate() {
	if PlayingMediaInfo.Playing {
		PlayingMediaInfo.Position += float64(time.Since(lastInterpolationCallTime).Microseconds()) / 1000000
	}
	lastInterpolationCallTime = time.Now()
}

func Start() {
	for {
		PlayingMediaInfo = GetCurrentPlayingMediaInfo()
		lastInterpolationCallTime = time.Now()
		time.Sleep(time.Second * 1)
	}
}

func ListPlayers() []string {
	conn, err := dbus.SessionBus()
	utils.CheckError(tracerr.Wrap(err))
	var names []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	utils.CheckError(tracerr.Wrap(err))

	var mprisNames []string
	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2") {
			mprisNames = append(mprisNames, name)
		}
	}
	return mprisNames
}

func GetCurrentPlayingMediaInfo() CurrentPlayingMediaInfo {
	switch runtime.GOOS {
	case "linux":
		players := ListPlayers()
		if len(players) == 0 {
			return CurrentPlayingMediaInfo{"No media", "", "", 0, 0, false}
		}
		p, err := mpris.NewPlayer(players[0])
		utils.CheckError(tracerr.Wrap(err))
		mediaPositionMicroseconds, err := p.Position()
		if err != nil {
			mediaPositionMicroseconds = 0
		}
		mediaPosition := float64(mediaPositionMicroseconds) / 1000000
		mediaMetadata, err := p.Metadata()
		utils.CheckError(tracerr.Wrap(err))
		mediaDurationMicroseconds, err := mediaMetadata.MPRISLength()
		utils.CheckError(tracerr.Wrap(err))
		mediaDuration := float64(mediaDurationMicroseconds) / 1000000
		mediaTitle, err := mediaMetadata.XESAMTitle()
		utils.CheckError(tracerr.Wrap(err))
		mediaAlbum, err := mediaMetadata.XESAMAlbum()
		utils.CheckError(tracerr.Wrap(err))
		mediaArtists, err := mediaMetadata.XESAMArtist()
		utils.CheckError(tracerr.Wrap(err))
		mediaArtist := ""
		if len(mediaArtists) != 0 {
			mediaArtist = mediaArtists[0]
		}
		utils.CheckError(tracerr.Wrap(err))
		status, err := p.PlaybackStatus()
		utils.CheckError(tracerr.Wrap(err))
		playing := false
		if status == mpris.PlaybackStatusPlaying {
			playing = true
		}
		return CurrentPlayingMediaInfo{mediaTitle, mediaAlbum, mediaArtist, mediaPosition, mediaDuration, playing}
	default:
		return CurrentPlayingMediaInfo{"Platform not supported", "Sorry", "", 0, 0, false}
	}
}
