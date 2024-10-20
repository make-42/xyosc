package audio

import (
	"fmt"
	"xyosc/config"
	"xyosc/utils"

	"github.com/gen2brain/malgo"
	"github.com/smallnest/ringbuffer"
)

var SampleRingBuffer *ringbuffer.RingBuffer
var SampleSizeInBytes uint32

const format = malgo.FormatF32

func Init() {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	SampleRingBuffer = ringbuffer.New(int(config.Config.RingBufferSize))
	deviceConfig.Capture.Format = format
	SampleSizeInBytes = uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
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

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = format
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = config.Config.SampleRate
	deviceConfig.Alsa.NoMMap = 1
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
	for {
	}
	//device.Uninit()
}
