package mpris

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

const timeFormat = "2006-01-02T15:04-07:00"

var ErrTypeNotParsable = errors.New("the given type is not as expected")

type PlaybackStatus string

const (
	PlaybackStatusPlaying PlaybackStatus = "Playing"
	PlaybackStatusPaused  PlaybackStatus = "Paused"
	PlaybackStatusStopped PlaybackStatus = "Stopped"
)

type LoopStatus string

const (
	LoopStatusNone     LoopStatus = "None"
	LoopStatusTrack    LoopStatus = "Track"
	LoopStatusPlaylist LoopStatus = "Playlist"
)

// Metadata represents the mpris-metadata
// see: https://www.freedesktop.org/wiki/Specifications/mpris-spec/metadata/
type Metadata map[string]dbus.Variant

// MPRISTrackID returns a unique identity for this track within the context of an MPRIS object (eg: tracklist).
func (md Metadata) MPRISTrackID() (dbus.ObjectPath, error) {
	vt := md["mpris:trackid"].Value()
	if vt == nil {
		return "", nil
	}

	v, ok := vt.(dbus.ObjectPath)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to dbus.ObjectPath: %w", vt, ErrTypeNotParsable)
	}

	return v, nil
}

// MPRISLength returns the duration of the track in microseconds.
func (md Metadata) MPRISLength() (int64, error) {
	vl := md["mpris:length"].Value()
	if vl == nil {
		return 0, nil
	}

	v, ok := vl.(int64)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to int64: %w", vl, ErrTypeNotParsable)
	}

	return v, nil
}

// MPRISArtURL returns the location of an image representing the track or album. Clients should not assume this will
// continue to exist when the media player stops giving out the URL.
func (md Metadata) MPRISArtURL() (string, error) {
	va := md["mpris:artUrl"].Value()
	if va == nil {
		return "", nil
	}

	v, ok := va.(string)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to string: %w", va, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMAlbum returns the album name.
func (md Metadata) XESAMAlbum() (string, error) {
	va := md["xesam:album"].Value()
	if va == nil {
		return "", nil
	}

	v, ok := va.(string)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to string: %w", va, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMAlbumArtist returns the album artist(s)
func (md Metadata) XESAMAlbumArtist() ([]string, error) {
	va := md["xesam:albumArtist"].Value()
	if va == nil {
		return nil, nil
	}

	v, ok := va.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", va, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMArtist returns the track artist(s).
func (md Metadata) XESAMArtist() ([]string, error) {
	va := md["xesam:artist"].Value()
	if va == nil {
		return nil, nil
	}

	v, ok := va.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", va, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMAsText returns the track lyrics.
func (md Metadata) XESAMAsText() (string, error) {
	vt := md["xesam:asText"].Value()
	if vt == nil {
		return "", nil
	}

	v, ok := vt.(string)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to string: %w", vt, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMAudioBPM returns the speed of the music, in beats per minute.
func (md Metadata) XESAMAudioBPM() (int, error) {
	va := md["xesam:audioBPM"].Value()
	if va == nil {
		return 0, nil
	}

	v, ok := va.(int)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to int: %w", va, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMAutoRating returns an automatically-generated rating, based on things such as how often it has been played.
// This should be in the range 0.0 to 1.0.
func (md Metadata) XESAMAutoRating() (float64, error) {
	va := md["xesam:autoRating"].Value()
	if va == nil {
		return 0, nil
	}

	v, ok := va.(float64)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to float64: %w", va, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMComment returns a (list of) freeform comment(s).
func (md Metadata) XESAMComment() ([]string, error) {
	vc := md["xesam:comment"].Value()
	if vc == nil {
		return nil, nil
	}

	v, ok := vc.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", vc, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMComposer returns the composer(s) of the track.
func (md Metadata) XESAMComposer() ([]string, error) {
	vc := md["xesam:composer"].Value()
	if vc == nil {
		return nil, nil
	}

	v, ok := vc.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", vc, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMContentCreated returns when the track was created. Usually only the year component will be useful.
func (md Metadata) XESAMContentCreated() (time.Time, error) {
	vc := md["xesam:contentCreated"].Value()
	if vc == nil {
		return time.Time{}, nil
	}

	vs, ok := vc.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("%T could not be parsed to string: %w", vc, ErrTypeNotParsable)
	}

	t, err := time.Parse(timeFormat, vs)
	if err != nil {
		return time.Time{}, fmt.Errorf("cound not parse time: %s: %w", err, ErrTypeNotParsable)
	}

	return t, nil
}

// XESAMDiscNumber returns the disc number on the album that this track is from.
func (md Metadata) XESAMDiscNumber() (int, error) {
	vn := md["xesam:discNumber"].Value()
	if vn == nil {
		return 0, nil
	}

	v, ok := vn.(int)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to int: %w", vn, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMFirstUsed returns when the track was first played.
func (md Metadata) XESAMFirstUsed() (time.Time, error) {
	vu := md["xesam:firstUsed"].Value()
	if vu == nil {
		return time.Time{}, nil
	}

	vs, ok := vu.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("%T could not be parsed to string: %w", vu, ErrTypeNotParsable)
	}

	t, err := time.Parse(timeFormat, vs)
	if err != nil {
		return time.Time{}, fmt.Errorf("cound not parse time: %s: %w", err, ErrTypeNotParsable)
	}

	return t, nil
}

// XESAMGenre returns the genre(s) of the track.
func (md Metadata) XESAMGenre() ([]string, error) {
	vg := md["xesam:genre"].Value()
	if vg == nil {
		return nil, nil
	}

	v, ok := vg.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", vg, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMLastUsed returns when the track was last played.
func (md Metadata) XESAMLastUsed() (time.Time, error) {
	vu := md["xesam:lastUsed"].Value()
	if vu == nil {
		return time.Time{}, nil
	}

	vs, ok := vu.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("%T could not be parsed to string: %w", vu, ErrTypeNotParsable)
	}

	t, err := time.Parse(timeFormat, vs)
	if err != nil {
		return time.Time{}, fmt.Errorf("cound not parse time: %s: %w", err, ErrTypeNotParsable)
	}

	return t, nil
}

// XESAMLyricist returns the lyricist(s) of the track.
func (md Metadata) XESAMLyricist() ([]string, error) {
	vl := md["xesam:lyricist"].Value()
	if vl == nil {
		return nil, nil
	}

	v, ok := vl.([]string)
	if !ok {
		return nil, fmt.Errorf("%T could not be parsed to []string: %w", vl, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMTitle returns the track title.
func (md Metadata) XESAMTitle() (string, error) {
	vt := md["xesam:title"].Value()
	if vt == nil {
		return "", nil
	}

	v, ok := vt.(string)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to string: %w", vt, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMTrackNumber returns the track number on the album disc.
func (md Metadata) XESAMTrackNumber() (int, error) {
	vn := md["xesam:trackNumber"].Value()
	if vn == nil {
		return 0, nil
	}

	v, ok := vn.(int)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to int: %w", vn, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMURL returns the location of the media file.
func (md Metadata) XESAMURL() (string, error) {
	vu := md["xesam:url"].Value()
	if vu == nil {
		return "", nil
	}

	v, ok := vu.(string)
	if !ok {
		return "", fmt.Errorf("%T could not be parsed to string: %w", vu, ErrTypeNotParsable)
	}

	return v, nil
}

// XESAMUseCount returns hte number of times the track has been played.
func (md Metadata) XESAMUseCount() (int, error) {
	vc := md["xesam:useCount"].Value()
	if vc == nil {
		return 0, nil
	}

	v, ok := vc.(int)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to int: %w", vc, ErrTypeNotParsable)
	}
	return v, nil
}

// XESAMUserRating returns a user-specified rating. This should be in the range 0.0 to 1.0.
func (md Metadata) XESAMUserRating() (float64, error) {
	vr := md["xesam:userRating"].Value()
	if vr == nil {
		return 0, nil
	}

	v, ok := vr.(float64)
	if !ok {
		return 0, fmt.Errorf("%T could not be parsed to float64: %w", vr, ErrTypeNotParsable)
	}
	return v, nil
}

// Find returns a generic representation of the requested value when present
func (md Metadata) Find(key string) (dbus.Variant, bool) {
	variant, found := md[key]
	return variant, found
}
