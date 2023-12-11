package translate

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/smilingpoplar/translate/google"
	"github.com/smilingpoplar/translate/util"
	"github.com/spf13/cobra"
)

type Translator interface {
	Translate(texts []string, toLang string) ([]string, error)
}

var (
	engine string
	tolang string
	proxy  string
)

func Main() int {
	var cmd = &cobra.Command{
		Short: "translate text to target language",
		Use: `translate "hello world"
  cat input.txt | translate > output.txt`,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := translate(args); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		},
	}
	cmd.Flags().StringVarP(&engine, "engine", "e", "google", "translate engine")
	cmd.Flags().StringVarP(&tolang, "tolang", "t", "zh-CN", "target language")
	cmd.Flags().StringVarP(&proxy, "proxy", "p", "", `http or socks5 proxy,
eg. http://127.0.0.1:7890 or socks5://127.0.0.1:7890`)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func translate(args []string) error {
	var cli Translator
	if engine == "google" {
		g := google.New()
		if err := util.SetProxy(proxy, g.Client); err != nil {
			return err
		}
		cli = g
	} else {
		return fmt.Errorf("unsupported engine: %s", engine)
	}

	var err error
	var reader io.Reader = os.Stdin
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
		return err
	}
	var result []string
	if result, err = cli.Translate(texts, tolang); err != nil {
		return err
	}
	if err = util.WriteLines(writer, result); err != nil {
		return err
	}

	return nil
}
