package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/smilingpoplar/translate/config"
	"github.com/smilingpoplar/translate/translator"
	"github.com/smilingpoplar/translate/util"
	"github.com/spf13/cobra"
)

const (
	kService = "service"
	kTolang  = "tolang"
	kEnvFile = "envfile"
	kProxy   = "proxy"
	KFixFile = "fixfile"
)

var (
	service string
	tolang  string
	envfile string
	proxy   string
	fixfile string
)

func main() {
	cmd := initCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "translate text to target language",
		Use: `translate "hello world"
  cat input.txt | translate > output.txt`,
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initEnv(); err != nil {
				return err
			}
			return translate(args)
		},
	}

	services := fmt.Sprintf("translate service, eg. %s", strings.Join(config.GetAllServiceNames(), ", "))
	cmd.Flags().StringVarP(&service, kService, "s", "google", services)
	cmd.Flags().StringVarP(&tolang, kTolang, "t", "zh-CN", "target language")
	cmd.Flags().StringVarP(&envfile, kEnvFile, "e", "", "env file, search .env upwards if not set")
	cmd.Flags().StringVarP(&proxy, kProxy, "p", "", "http or socks5 proxy,\n eg. http://127.0.0.1:7890 or socks5://127.0.0.1:7890")
	cmd.Flags().StringVarP(&fixfile, KFixFile, "f", "", "csv file to fix translation")

	return cmd
}

func initEnv() error {
	filename := envfile
	if filename == "" {
		filename = ".env"
	}

	path, err := util.FileExistsInParentDirs(filename)
	if err != nil { // 文件不存在
		if envfile != "" {
			return fmt.Errorf("error envfile: %w", err)
		}
		return nil
	}

	if err := godotenv.Load(path); err != nil {
		if envfile != "" {
			return fmt.Errorf("error loading envfile: %w", err)
		}
	}
	return nil
}

func translate(args []string) error {
	var trans translator.Translator
	var err error
	if trans, err = translator.GetTranslator(service, proxy); err != nil {
		return err
	}
	var fixes []util.FixTransform
	if fixes, err = util.LoadTranslationFixes(fixfile); err != nil {
		return err
	}

	var reader io.Reader = os.Stdin
	if len(args) == 0 { // 从os.Stdin读取要翻译的文本
		if util.IsTerminal() {
			return translateInTerminal(trans, fixes)
		}
	} else { // 翻译命令行参数
		reader = strings.NewReader(strings.Join(args, "\n"))
	}
	var writer io.Writer = os.Stdout

	o, ok := trans.(translator.TranslationObserver)
	if ok { // 收到分组响应后立即输出
		o.OnTranslated(func(result []string) error {
			return OnTranslated(writer, result, fixes)
		})
	}
	var texts, result []string
	if texts, err = util.ReadLines(reader); err != nil {
		return err
	}
	if result, err = trans.Translate(texts, tolang); err != nil {
		return err
	}
	if !ok { // 收到全部响应后再输出
		if err = OnTranslated(writer, result, fixes); err != nil {
			return err
		}
	}

	return nil
}

func translateInTerminal(trans translator.Translator, fixes []util.FixTransform) error {
	fmt.Println("Input texts to be translated... <Ctrl-D> to finish.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		result, err := trans.Translate([]string{text}, tolang)
		if err != nil {
			return err
		}
		util.ApplyTranslationFixes(result, fixes)
		fmt.Fprintln(os.Stdout, result[0])
	}
	return scanner.Err()
}

func OnTranslated(writer io.Writer, result []string, fixes []util.FixTransform) error {
	util.ApplyTranslationFixes(result, fixes)
	return util.WriteLines(writer, result)
}
