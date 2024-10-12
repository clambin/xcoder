package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"log/slog"
	"os"
)

var asJSON = flag.Bool("json", false, "dump stats as json")

func main() {
	flag.Parse()
	p := ffmpeg.Processor{Logger: slog.Default()}

	for _, arg := range flag.Args() {
		probe, err := p.Scan(context.Background(), arg)
		if err != nil {
			panic(err)
		}

		if *asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(probe)
		} else {
			fmt.Printf("%s: codec:%s bitrate:%s height:%d width:%d\n",
				arg,
				probe.VideoCodec,
				ffmpeg.Bits(probe.BitRate).Format(2),
				probe.Height,
				probe.Width,
			)
		}
	}
}
