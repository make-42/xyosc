package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/utils"
)

var SampleRingBufferUnsafe []float32
var SampleSizeInBytes uint32
var WriteHeadPosition uint32

const format = malgo.FormatF32

func Init() {
	SampleRingBufferUnsafe = make([]float32, int(config.Config.Buffers.RingBufferSize))
	SampleSizeInBytes = uint32(malgo.SampleSizeInBytes(format))
	WriteHeadPosition = 0
}

func Start() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	utils.CheckError(tracerr.Wrap(err))
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// Capture devices.
	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	overrideCaptureDeviceIndex := -1
	fmt.Println("Capture Devices")
	for i, info := range infos {
		e := "ok"
		full, err := ctx.DeviceInfo(malgo.Capture, info.ID, malgo.Shared)
		if err != nil {
			e = err.Error()
		}
		fmt.Printf("    %d: %v, %s, [%s], formats: %+v\n",
			i, info.ID, info.Name(), e, full.Formats)
		if info.Name() == config.Config.Audio.CaptureDeviceMatchName && config.Config.Audio.CaptureDeviceMatchName != "" && (config.Config.Audio.CaptureDeviceMatchSampleRate == 0 || config.Config.Audio.CaptureDeviceMatchSampleRate == int(full.Formats[0].SampleRate)) {
			overrideCaptureDeviceIndex = i
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = format
	deviceConfig.Capture.Channels = 2
	if overrideCaptureDeviceIndex != -1 {
		deviceConfig.Capture.DeviceID = infos[overrideCaptureDeviceIndex].ID.Pointer()
	} else {
		deviceConfig.Capture.DeviceID = infos[config.Config.Audio.CaptureDeviceMatchIndex].ID.Pointer()
	}
	deviceConfig.PerformanceProfile = malgo.LowLatency
	deviceConfig.SampleRate = config.Config.Audio.SampleRate
	deviceConfig.Alsa.NoMMap = 1
	deviceConfig.PeriodSizeInFrames = config.Config.Buffers.AudioCaptureBufferSize

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		buf := bytes.NewReader(pSample)
		var AX float32
		var AY float32
		i := 0
		for {
			if binary.Read(buf, binary.NativeEndian, &AX) != nil || binary.Read(buf, binary.NativeEndian, &AY) != nil {
				break
			}
			SampleRingBufferUnsafe[int(WriteHeadPosition)+i*2] = AX
			SampleRingBufferUnsafe[int(WriteHeadPosition)+i*2+1] = AY
			i++
		}
		WriteHeadPosition = (WriteHeadPosition + uint32(len(pSample))/4) % config.Config.Buffers.RingBufferSize
	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)

	utils.CheckError(tracerr.Wrap(err))

	err = device.Start()

	utils.CheckError(tracerr.Wrap(err))
	select {}
	//device.Uninit()
}
