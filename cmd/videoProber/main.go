package main

import (
	"context"
	"fmt"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"log/slog"
	"os"
)

func main() {
	p := ffmpeg.Processor{Logger: slog.Default()}
	for _, arg := range os.Args[1:] {
		probe, err := p.Probe(context.Background(), arg)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: codec:%s bitrate:%.1fmbps height:%d\n",
			arg,
			probe.VideoCodec(),
			float64(probe.BitRate())/1024/1024,
			probe.Height(),
		)
	}
}
