package mpris

import (
	"context"
	"fmt"
	"github.com/godbus/dbus/v5"
)

const (
	playerObjectPath             = "/org/mpris/MediaPlayer2"
	playerInterface              = "org.mpris.MediaPlayer2.Player"
	playerNextMethod             = playerInterface + ".Next"
	playerPreviousMethod         = playerInterface + ".Previous"
	playerPauseMethod            = playerInterface + ".Pause"
	playerPlayPauseMethod        = playerInterface + ".PlayPause"
	playerStopMethod             = playerInterface + ".Stop"
	playerPlayMethod             = playerInterface + ".Play"
	playerSeekMethod             = playerInterface + ".SeekTo"
	playerSetPositionMethod      = playerInterface + ".SetPosition"
	playerOpenURIMethod          = playerInterface + ".OpenURI"
	playerPlaybackStatusProperty = playerInterface + ".PlaybackStatus"
	playerLoopStatusProperty     = playerInterface + ".LoopStatus"
	playerRateProperty           = playerInterface + ".Rate"
	playerShuffleProperty        = playerInterface + ".Shuffle"
	playerMetadataProperty       = playerInterface + ".Metadata"
	playerVolumeProperty         = playerInterface + ".Volume"
	playerPositionProperty       = playerInterface + ".Position"
	playerMinimumRateProperty    = playerInterface + ".MinimumRate"
	playerMaximumRateProperty    = playerInterface + ".MaximumRate"
	playerCanGoNextProperty      = playerInterface + ".CanGoNext"
	playerCanGoPreviousProperty  = playerInterface + ".CanGoPrevious"
	playerCanPlayProperty        = playerInterface + ".CanPlay"
	playerCanPauseProperty       = playerInterface + ".CanPause"
	playerCanSeekProperty        = playerInterface + ".CanSeek"
	playerCanControlProperty     = playerInterface + ".CanControl"
	signalNameSeeked             = "org.mpris.MediaPlayer2.Player.Seeked"
)

var dbusSessionBus = dbus.SessionBus

//go:generate moq -out dbus-conn_moq_test.go . dbusConn
type dbusConn interface {
	Object(string, dbus.ObjectPath) dbusBusObject
	AddMatchSignal(...dbus.MatchOption) error
	Signal(ch chan<- *dbus.Signal)
	Close() error
}

//go:generate moq -out dbus-bus-object_moq_test.go . dbusBusObject
type dbusBusObject interface {
	Call(method string, flags dbus.Flags, args ...interface{}) dbusCall
	GetProperty(p string) (v dbus.Variant, e error)
	SetProperty(p string, v interface{}) error
}

//go:generate moq -out dbus-call_moq_test.go . dbusCall
type dbusCall interface {
	Store(retvalues ...interface{}) error
}

// Player is a implementation of dbus org.mpris.MediaPlayer2.Player. see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html
// Use NewPlayer to create a new instance with a connected session-bus via dbus.SessionBus.
// Use NewPlayerWithConnection when you want to use a self-configured dbus.Conn
type Player struct {
	name       string
	connection dbusConn
}

// NewPlayer returns a new Player which is already connected to session-bus via dbus.SessionBus.
// Don't forget to Player.Close() the player after use.
func NewPlayer(name string) (Player, error) {
	connection, err := dbusSessionBus()
	if err != nil {
		return Player{}, fmt.Errorf("failed to connect to session-bus: %w", err)
	}

	return Player{
		name: name,
		connection: &dbusConnWrapper{
			conn: connection,
		},
	}, nil
}

// NewPlayerWithConnection returns a new Player with the given name and connection.
// Deprecated: NewPlayerWithConnection will be removed in future (v2.X.X).
// Plain Struct initialization should be used instead.
// Private fields will be public.
func NewPlayerWithConnection(name string, connection *dbus.Conn) Player {
	return Player{
		name: name,
		connection: &dbusConnWrapper{
			conn: connection,
		},
	}
}

// Close closes the dbus connection.
func (p Player) Close() error {
	err := p.connection.Close()
	if err != nil {
		return fmt.Errorf("failed to close dbus connection: %w", err)
	}

	return nil
}

// Next skips to the next track in the tracklist.
// If there is no next track (and endless playback and track repeat are both off), stop playback.
// If playback is paused or stopped, it remains that way.
// If CanGoNext is false, attempting to call this method should have no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p Player) Next() {
	p.connection.Object(p.name, playerObjectPath).Call(playerNextMethod, 0)
}

// Previous skips to the previous track in the tracklist.
// If there is no previous track (and endless playback and track repeat are both off), stop playback.
// If playback is paused or stopped, it remains that way.
// If CanGoPrevious is false, attempting to call this method should have no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p Player) Previous() {
	p.connection.Object(p.name, playerObjectPath).Call(playerPreviousMethod, 0)
}

// Pause pauses playback.
// If playback is already paused, this has no effect.
// Calling Play after this should cause playback to start again from the same position.
// If CanPause is false, attempting to call this method should have no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p Player) Pause() {
	p.connection.Object(p.name, playerObjectPath).Call(playerPauseMethod, 0)
}

// Play starts or resumes playback.
// If already playing, this has no effect.
// If paused, playback resumes from the current position.
// If there is no track to play, this has no effect.
// If CanPlay is false, attempting to call this method should have no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p Player) Play() {
	p.connection.Object(p.name, playerObjectPath).Call(playerPlayMethod, 0)
}

// PlayPause pauses playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// If CanPause is false, attempting to call this method should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p Player) PlayPause() {
	p.connection.Object(p.name, playerObjectPath).Call(playerPlayPauseMethod, 0)
}

// Stop stops playback.
// If playback is already stopped, this has no effect.
// Calling Play after this should cause playback to start again from the beginning of the track.
// If CanControl is false, attempting to call this method should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p Player) Stop() {
	p.connection.Object(p.name, playerObjectPath).Call(playerStopMethod, 0)
}

// SeekTo seeks forward in the current track by the specified number of microseconds.
// Parameters:
// - offset (The number of microseconds to seek forward.)
// A negative value seeks back. If this would mean seeking back further than the start of the track, the position is set to 0.
// If the value passed in would mean seeking beyond the end of the track, acts like a call to Next.
// If the CanSeek property is false, this has no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p Player) SeekTo(offset int64) {
	p.connection.Object(p.name, playerObjectPath).Call(playerSeekMethod, 0, offset)
}

// OpenURI opens the Uri given as an argument
// Parameters:
// - uri (Uri of the track to load. Its uri scheme should be an element of the org.mpris.MediaPlayer2.SupportedUriSchemes property and the mime-type should match one of the elements of the org.mpris.MediaPlayer2.SupportedMimeTypes.)
// If the playback is stopped, starts playing
// If the uri scheme or the mime-type of the uri to open is not supported, this method does nothing and may raise an error. In particular, if the list of available uri schemes is empty, this method may not be implemented.
// Clients should not assume that the Uri has been opened as soon as this method returns. They should wait until the mpris:trackid field in the Metadata property changes.
// If the media player implements the TrackList interface, then the opened track should be made part of the tracklist, the org.mpris.MediaPlayer2.TrackList.TrackAdded or org.mpris.MediaPlayer2.TrackList.TrackListReplaced signal should be fired, as well as the org.freedesktop.DBus.Properties.PropertiesChanged signal on the tracklist interface.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:OpenUri
func (p Player) OpenURI(uri string) {
	p.connection.Object(p.name, playerObjectPath).Call(playerOpenURIMethod, 0, uri)
}

// PlaybackStatus returns the current playback status.
// May be "Playing" as PlaybackStatusPlaying, "Paused" as PlaybackStatusPaused or "Stopped" as PlaybackStatusStopped.
// https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:PlaybackStatus
func (p Player) PlaybackStatus() (PlaybackStatus, error) {
	v, err := p.getProperty(playerPlaybackStatusProperty)
	if err != nil {
		return "", err
	}
	return PlaybackStatus(v.Value().(string)), nil
}

// LoopStatus returns the current loop / repeat status
// May be:
// "None" as LoopStatusNone if the playback will stop when there are no more tracks to play
// "Track" as LoopStatusTrack if the current track will start again from the beginning once it has finished playing
// "Playlist" as LoopStatusPlaylist if the playback loops through a list of tracks
// If CanControl is false, attempting to set this property (SetLoopStatus) should have no effect and raise an error.
// https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:LoopStatus
func (p Player) LoopStatus() (LoopStatus, error) {
	v, err := p.getProperty(playerLoopStatusProperty)
	if err != nil {
		return "", err
	}
	return LoopStatus(v.Value().(string)), nil
}

// SetLoopStatus sets the current loop / repeat status
// May be:
// "None" as LoopStatusNone if the playback will stop when there are no more tracks to play
// "Track" as LoopStatusTrack if the current track will start again from the beginning once it has finished playing
// "Playlist" as LoopStatusPlaylist if the playback loops through a list of tracks
// If CanControl is false, attempting to set this property (SetLoopStatus) should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:LoopStatus
func (p Player) SetLoopStatus(status LoopStatus) error {
	return p.setProperty(playerLoopStatusProperty, string(status))
}

// Rate return the current playback rate.
// The value must fall in the range described by MinimumRate and MaximumRate, and must not be 0.0. If playback is paused, the PlaybackStatus property should be used to indicate this. A value of 0.0 should not be set by the client. If it is, the media player should act as though Pause was called.
// If the media player has no ability to play at speeds other than the normal playback rate, this must still be implemented, and must return 1.0. The MinimumRate and MaximumRate properties must also be set to 1.0.
// Not all values may be accepted by the media player. It is left to media player implementations to decide how to deal with values they cannot use; they may either ignore them or pick a "best fit" value. Clients are recommended to only use sensible fractions or multiples of 1 (eg: 0.5, 0.25, 1.5, 2.0, etc).
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Rate
func (p Player) Rate() (float64, error) {
	v, err := p.getProperty(playerRateProperty)
	if err != nil {
		return 0, err
	}
	return v.Value().(float64), nil
}

// SetRate sets the current playback rate.
// The value must fall in the range described by MinimumRate and MaximumRate, and must not be 0.0. If playback is paused, the PlaybackStatus property should be used to indicate this. A value of 0.0 should not be set by the client. If it is, the media player should act as though Pause was called.
// If the media player has no ability to play at speeds other than the normal playback rate, this must still be implemented, and must return 1.0. The MinimumRate and MaximumRate properties must also be set to 1.0.
// Not all values may be accepted by the media player. It is left to media player implementations to decide how to deal with values they cannot use; they may either ignore them or pick a "best fit" value. Clients are recommended to only use sensible fractions or multiples of 1 (eg: 0.5, 0.25, 1.5, 2.0, etc).
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Rate
func (p Player) SetRate(rate float64) error {
	return p.setProperty(playerRateProperty, rate)
}

// Shuffle returns a value of false indicates that playback is progressing linearly through a playlist, while true means playback is progressing through a playlist in some other order.
// If CanControl is false, attempting to set this property should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Shuffle
func (p Player) Shuffle() (bool, error) {
	v, err := p.getProperty(playerShuffleProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// SetShuffle set a value of false indicates that playback is progressing linearly through a playlist, while true means playback is progressing through a playlist in some other order.
// If CanControl is false, attempting to set this property should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Shuffle
func (p Player) SetShuffle(shuffle bool) error {
	return p.setProperty(playerShuffleProperty, shuffle)
}

// Metadata of the current element.
// If there is a current track, this must have a "mpris:trackid" entry (of D-Bus type "o") at the very least, which contains a D-Bus path that uniquely identifies this track.
// See the type documentation for more details.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Metadata
func (p Player) Metadata() (Metadata, error) {
	v, err := p.getProperty(playerMetadataProperty)
	if err != nil {
		return nil, err
	}
	return v.Value().(map[string]dbus.Variant), nil
}

// Volume returns the volume level.
// When setting, if a negative value is passed, the volume should be set to 0.0.
// If CanControl is false, attempting to set this property should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Volume
func (p Player) Volume() (float64, error) {
	v, err := p.getProperty(playerVolumeProperty)
	if err != nil {
		return 0, err
	}
	return v.Value().(float64), nil
}

// SetVolume sets the volume level.
// When setting, if a negative value is passed, the volume should be set to 0.0.
// If CanControl is false, attempting to set this property should have no effect and raise an error.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Volume
func (p Player) SetVolume(volume float64) error {
	return p.setProperty(playerVolumeProperty, volume)
}

// Position returns current track position in microseconds, between 0 and the 'mpris:length' metadata entry (see Metadata).
// Note: If the media player allows it, the current playback position can be changed either the SetPosition method or the Seek method on this interface. If this is not the case, the CanSeek property is false, and setting this property has no effect and can raise an error.
// If the playback progresses in a way that is inconsistent with the Rate property, the Seeked signal is emitted.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:Position
func (p Player) Position() (int64, error) {
	v, err := p.getProperty(playerPositionProperty)
	if err != nil {
		return 0, err
	}
	return v.Value().(int64), nil
}

// SetPosition Sets the current track position in microseconds.
// Parameters:
// - trackID (The currently playing track's identifier. If this does not match the id of the currently-playing track, the call is ignored as "stale".)
// - position (Track position in microseconds. This must be between 0 and <track_length>.)
// If this does not match the id of the currently-playing track, the call is ignored as "stale".
// If the Position argument is less than 0, do nothing.
// If the Position argument is greater than the track length, do nothing.
// If the CanSeek property is false, this has no effect.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:SetPosition
func (p Player) SetPosition(trackID dbus.ObjectPath, position int64) {
	p.connection.Object(p.name, playerObjectPath).Call(playerSetPositionMethod, 0, trackID, position)
}

// MinimumRate returns the minimum value which the Rate property can take. Clients should not attempt to set the Rate property below this value.
// Note that even if this value is 0.0 or negative, clients should not attempt to set the Rate property to 0.0.
// This value should always be 1.0 or less.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:MinimumRate
func (p Player) MinimumRate() (float64, error) {
	v, err := p.getProperty(playerMinimumRateProperty)
	if err != nil {
		return 0, err
	}
	return v.Value().(float64), nil
}

// MaximumRate returns the maximum value which the Rate property can take. Clients should not attempt to set the Rate property above this value.
// This value should always be 1.0 or greater.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:MaximumRate
func (p Player) MaximumRate() (float64, error) {
	v, err := p.getProperty(playerMaximumRateProperty)
	if err != nil {
		return 0, err
	}
	return v.Value().(float64), nil
}

// CanGoNext returns true whether the client can call the Next method on this interface and expect the current track to change.
// If it is unknown whether a call to Next will be successful (for example, when streaming tracks), this property should be set to true.
// If CanControl is false, this property should also be false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanGoNext
func (p Player) CanGoNext() (bool, error) {
	v, err := p.getProperty(playerCanGoNextProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// CanGoPrevious returns true whether the client can call the Previous method on this interface and expect the current track to change.
// If it is unknown whether a call to Previous will be successful (for example, when streaming tracks), this property should be set to true.
// If CanControl is false, this property should also be false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanGoPrevious
func (p Player) CanGoPrevious() (bool, error) {
	v, err := p.getProperty(playerCanGoPreviousProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// CanPlay returns true whether playback can be started using Play or PlayPause.
// Note that this is related to whether there is a "current track": the value should not depend on whether the track is currently paused or playing. In fact, if a track is currently playing (and CanControl is true), this should be true.
// If CanControl is false, this property should also be false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanPlay
func (p Player) CanPlay() (bool, error) {
	v, err := p.getProperty(playerCanPlayProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// CanPause returns true whether playback can be paused using Pause or PlayPause.
// Note that this is an intrinsic property of the current track: its value should not depend on whether the track is currently paused or playing. In fact, if playback is currently paused (and CanControl is true), this should be true.
// If CanControl is false, this property should also be false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanPause
func (p Player) CanPause() (bool, error) {
	v, err := p.getProperty(playerCanPauseProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// CanSeek returns true whether the client can control the playback position using Seek and SetPosition. This may be different for different tracks.
// If CanControl is false, this property should also be false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanSeek
func (p Player) CanSeek() (bool, error) {
	v, err := p.getProperty(playerCanSeekProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// CanControl is true whether the media player may be controlled over this interface.
// This property is not expected to change, as it describes an intrinsic capability of the implementation.
// If this is false, clients should assume that all properties on this interface are read-only (and will raise errors if writing to them is attempted), no methods are implemented and all other properties starting with "Can" are also false.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Property:CanControl
func (p Player) CanControl() (bool, error) {
	v, err := p.getProperty(playerCanControlProperty)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

// Seeked indicates that the track position has changed in a way that is inconsistant with the current playing state.
// When this signal is not received, clients should assume that:
// - When playing, the position progresses according to the rate property.
// - When paused, it remains constant.
// This signal does not need to be emitted when playback starts or when the track changes, unless the track is starting
// at an unexpected position. An expected position would be the last known one when going from Paused to Playing, and 0
// when going from Stopped to Playing.
// see: https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html#Signal:Seeked
func (p Player) Seeked(ctx context.Context) (<-chan int, error) {
	err := p.connection.AddMatchSignal(dbus.WithMatchSender("org.mpris.MediaPlayer2.vlc"))
	if err != nil {
		return nil, fmt.Errorf("failed to add signal match option: %w", err)
	}

	positions := make(chan int)
	errs := make(chan error)
	go func() {
		defer func() {
			close(positions)
			close(errs)
		}()

		signals := make(chan *dbus.Signal)
		defer close(signals)
		p.connection.Signal(signals)

	collectPositions:
		for {
			select {
			case sig := <-signals:
				if sig.Name != signalNameSeeked || // irrelevant signal
					len(sig.Body) != 1 { // invalid event
					continue
				}
				micros, ok := sig.Body[0].(int64)
				if !ok { // broken signal
					continue
				}
				positions <- int(micros)
			case <-ctx.Done():
				break collectPositions
			}
		}
	}()

	return positions, nil
}

func (p Player) getProperty(property string) (dbus.Variant, error) {
	v, err := p.connection.Object(p.name, playerObjectPath).GetProperty(property)
	if err != nil {
		return dbus.Variant{}, fmt.Errorf("failed to get property %q: %w", property, err)
	}

	return v, nil
}

func (p Player) setProperty(property string, value interface{}) error {
	err := p.connection.Object(p.name, playerObjectPath).SetProperty(property, dbus.MakeVariant(value))
	if err != nil {
		return fmt.Errorf("failed to set property %q: %w", property, err)
	}

	return nil
}
