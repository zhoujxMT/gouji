package frame

const (
	MODE_RELEASE = iota
	MODE_DEBUG
)

var (
	mode int
)

type Config struct {
	Mode int
}

func loadConfig() {
	if args.Config == nil {
		return
	}
	mode = args.Config.Mode
}

func GetMode() int {
	return mode
}

func SetMode(mode_ int) {
	mode = mode_
}
