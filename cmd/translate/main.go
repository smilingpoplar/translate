package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/smilingpoplar/translate/translator"
	"github.com/smilingpoplar/translate/translator/google"
	"github.com/smilingpoplar/translate/translator/openai"
	"github.com/smilingpoplar/translate/util"
	"github.com/spf13/cobra"
)

var (
	engine string
	tolang string
	proxy  string
)

var (
	oaiAPIKey  string
	oaiAPIBase string
)

func main() {
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
	cmd.Flags().StringVarP(&engine, "engine", "e", "google", `translate engine,
eg. google, openai`)
	cmd.Flags().StringVarP(&tolang, "tolang", "t", "zh-CN", "target language")
	cmd.Flags().StringVarP(&proxy, "proxy", "p", "", `http or socks5 proxy,
eg. http://127.0.0.1:7890 or socks5://127.0.0.1:7890`)
	cmd.Flags().StringVarP(&oaiAPIKey, "api-key", "k", "", "required: openai")
	cmd.Flags().StringVarP(&oaiAPIBase, "api-base", "b", "", fmt.Sprintf("optional: openai (default %q)", openai.BaseURL))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func translate(args []string) error {
	var trans translator.Translator
	var err error
	if trans, err = getTranslator(); err != nil {
		return err
	}

	var reader io.Reader = os.Stdin
	if len(args) == 0 { // 从os.Stdin读取要翻译的文本
		if util.IsTerminal() {
			return translateInTerminal(trans)
		}
	} else { // 翻译命令行参数
		reader = strings.NewReader(strings.Join(args, "\n"))
	}
	var writer io.Writer = os.Stdout

	o, ok := trans.(translator.TranslationObserver)
	if ok { // 收到分组响应后立即输出
		o.OnTranslated(func(result []string) error {
			return util.WriteLines(writer, result)
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
		if err = util.WriteLines(writer, result); err != nil {
			return err
		}
	}

	return nil
}

func translateInTerminal(trans translator.Translator) error {
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
		fmt.Fprintln(os.Stdout, result[0])
	}
	return scanner.Err()
}

func getTranslator() (translator.Translator, error) {
	var trans translator.Translator
	var err error
	if engine == "google" {
		trans, err = google.NewWithProxy(proxy)
	} else if engine == "openai" {
		trans, err = getTranslatorOpenAI()
	} else {
		err = fmt.Errorf("unsupported engine: %s", engine)
	}
	return trans, err
}

func getTranslatorOpenAI() (translator.Translator, error) {
	API_KEY, BASE_URL := "OPENAI_API_KEY", "OPENAI_BASE_URL"
	// 先从命令行获取，再从环境变量获取
	if oaiAPIKey == "" {
		oaiAPIKey = os.Getenv(API_KEY)
	}
	if oaiAPIBase == "" {
		oaiAPIBase = os.Getenv(BASE_URL)
	}

	if oaiAPIKey == "" {
		msg := fmt.Sprintf("%s is not set, set it with `export %s=YOUR_API_KEY`", API_KEY, API_KEY)
		if oaiAPIBase == "" {
			msg += fmt.Sprintf("\n%s default %q, set it with `export %s=YOUR_BASE_URL`", BASE_URL, openai.BaseURL, BASE_URL)
		}
		return nil, fmt.Errorf(msg)
	}
	return openai.NewWithProxy(oaiAPIKey, oaiAPIBase, proxy)
}
