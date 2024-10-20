package main

import (
	"encoding/binary"
	"xyosc/audio"
	"xyosc/config"
	"xyosc/fonts"
	"xyosc/media"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	config.Init()
	audio.Init()
	go audio.Start()
	go media.Start()

	scale := min(config.Config.WindowWidth, config.Config.WindowHeight) / 2
	rl.InitWindow(config.Config.WindowWidth, config.Config.WindowHeight, "xyosc")
	defer rl.CloseWindow()
	rl.SetWindowOpacity(config.Config.WindowOpacity)
	rl.SetConfigFlags(rl.FlagWindowTransparent)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.SetWindowState(rl.FlagWindowUndecorated)
	rl.SetTargetFPS(config.Config.TargetFPS)
	rl.SetWindowPosition(rl.GetMonitorWidth(rl.GetCurrentMonitor())/2, rl.GetMonitorHeight(rl.GetCurrentMonitor())/2)
	var AX float32
	var AY float32
	var BX float32
	var BY float32

	fonts.Init()
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
		for i := uint32(0); i < config.Config.ReadBufferSize/audio.SampleSizeInBytes/2; i++ {
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BY)
			fAX := float32(AX) * config.Config.Gain * float32(scale)
			fAY := -float32(AY) * config.Config.Gain * float32(scale)
			fBX := float32(BX) * config.Config.Gain * float32(scale)
			fBY := -float32(BY) * config.Config.Gain * float32(scale)
			rl.DrawLineEx(rl.NewVector2(float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowWidth/2)+fAY), rl.NewVector2(float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowWidth/2)+fBY), config.Config.LineThickness, config.AccentColor)
			AX = BX
			AY = BY
		}
		if config.Config.FPSCounter {
			rl.DrawFPS(16, config.Config.WindowHeight)
		}
		rl.DrawTextEx(fonts.FontIosevka32, media.PlayingMediaInfo.Artist+" - "+media.PlayingMediaInfo.Title, rl.NewVector2(16, 16), 32, 2, config.AccentColor)
		rl.DrawTextEx(fonts.FontIosevka16, media.PlayingMediaInfo.Album, rl.NewVector2(16, 48), 16, 1, config.ThirdColor)
		rl.DrawTextEx(fonts.FontIosevka32, media.FmtDuration(media.PlayingMediaInfo.Position)+" / "+media.FmtDuration(media.PlayingMediaInfo.Duration), rl.NewVector2(16, 64), 32, 2, config.AccentColor)
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
