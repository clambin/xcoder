package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/clambin/videoConvertor/internal/convertor"
	"log/slog"
	"math"
	"os"
	"time"
)

var asJSON = flag.Bool("json", false, "dump stats as json")

func main() {
	flag.Parse()
	p := convertor.Processor{Logger: slog.Default()}

	for _, arg := range flag.Args() {
		stats, err := p.Scan(context.Background(), arg)
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
				convertor.Bits(stats.BitRate).Format(1),
				stats.Height,
				stats.Width,
				(time.Duration(math.Round(stats.Duration.Seconds())) * time.Second).String(),
			)
		}
	}
}
