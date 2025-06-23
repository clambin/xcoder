package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
)

var asJSON = flag.Bool("json", false, "dump stats as json")

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		stats, err := ffmpeg.Probe(arg)
		if err != nil {
			panic(err)
		}

		if *asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(stats)
		} else {
			fmt.Printf("%s: codec:%s bitrate:%s height:%d width:%d duration: %6s\n",
				arg,
				stats.VideoCodec,
				ffmpeg.Bits(stats.BitRate).Format(1),
				stats.Height,
				stats.Width,
				(time.Duration(math.Round(stats.Duration.Seconds())) * time.Second).String(),
			)
		}
	}
}
