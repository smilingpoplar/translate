package translate

import (
	"flag"
	"fmt"
	"os"

	"github.com/smilingpoplar/translate/google"
)

type Translator interface {
	Translate(text string) (string, error)
}

func Main() int {
	engine := flag.String("engine", "google", "translate engine")
	flag.Parse()

	var cli Translator
	if *engine == "google" {
		cli = google.New()
	} else {
		fmt.Fprintln(os.Stderr, "not support engine:", *engine)
		return 1
	}

	text := flag.Arg(0)
	translated, err := cli.Translate(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, "err translate:", err)
		return 1
	}
	fmt.Println(translated)

	return 0
}
