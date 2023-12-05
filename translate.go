package translate

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/smilingpoplar/translate/google"
	"github.com/smilingpoplar/translate/util"
)

type Translator interface {
	Translate(texts []string, toLang string) ([]string, error)
}

func Main() int {
	engine := flag.String("engine", "google", "translate engine")
	tolang := flag.String("tolang", "zh-CN", "target language")
	proxy := flag.String("proxy", "", "http|socks5 proxy, e.g. http://127.0.0.1:7890 or socks5://127.0.0.1:7890")
	flag.Parse()

	var cli Translator
	if *engine == "google" {
		g := google.New()
		if err := util.SetProxy(*proxy, g.Client); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		cli = g
	} else {
		fmt.Fprintln(os.Stderr, "not support engine:", *engine)
		return 1
	}

	var err error
	var reader io.Reader = os.Stdin
	args := flag.Args()
	if len(args) > 0 { // 翻译命令行参数
		reader = strings.NewReader(strings.Join(args, "\n"))
	} else { // 从os.Stdin读取要翻译的文本
		if util.IsTerminal() { // 输出命令行提示
			fmt.Println("Input texts to be translated... <Ctrl-D> to finish.")
		}
	}
	var writer io.Writer = os.Stdout

	var texts []string
	if texts, err = util.ReadLines(reader); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var result []string
	if result, err = cli.Translate(texts, *tolang); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err = util.WriteLines(writer, result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
