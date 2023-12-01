package translate

import (
	"flag"
	"fmt"
	"os"

	"github.com/smilingpoplar/translate/google"
)

type Translator interface {
	Translate(texts []string) ([]string, error)
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

	texts := flag.Args()
	if len(texts) == 0 {
		fmt.Fprintln(os.Stderr, "no text to translate")
		return 1
	}

	result, err := cli.Translate(texts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "err translate:", err)
		return 1
	}
	for _, line := range result {
		fmt.Println(line)
	}

	return 0
}
