package audio

import (
	"fmt"
	"os"
	"xyosc/config"
	"xyosc/utils"

	"github.com/gen2brain/malgo"
	"github.com/smallnest/ringbuffer"
)

var SampleRingBuffer *ringbuffer.RingBuffer
var SampleSizeInBytes uint32

const format = malgo.FormatF32

func Init() {
	SampleRingBuffer = ringbuffer.New(int(config.Config.RingBufferSize)).SetBlocking(true)
	SampleSizeInBytes = uint32(malgo.SampleSizeInBytes(format))
}

func Start() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	utils.CheckError(err)
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
		if info.Name() == config.Config.CaptureDeviceName && config.Config.CaptureDeviceName != "" && (config.Config.CaptureDeviceSampleRate == 0 || config.Config.CaptureDeviceSampleRate == int(full.Formats[0].SampleRate)) {
			overrideCaptureDeviceIndex = i
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = format
	deviceConfig.Capture.Channels = 2
	if overrideCaptureDeviceIndex != -1 {
		deviceConfig.Capture.DeviceID = infos[overrideCaptureDeviceIndex].ID.Pointer()
	} else {
		deviceConfig.Capture.DeviceID = infos[config.Config.CaptureDeviceIndex].ID.Pointer()
	}
	deviceConfig.PerformanceProfile = malgo.LowLatency
	deviceConfig.SampleRate = config.Config.SampleRate
	deviceConfig.Alsa.NoMMap = 1
	deviceConfig.PeriodSizeInFrames = config.Config.AudioCaptureBufferSize

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		SampleRingBuffer.Write(pSample)
	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)

	utils.CheckError(err)

	err = device.Start()

	utils.CheckError(err)
	select {}
	//device.Uninit()
}
