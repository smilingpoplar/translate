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
	kService   = "service"
	kTolang    = "tolang"
	kEnvFile   = "envfile"
	kProxy     = "proxy"
	KGlossFile = "glossfile"
	kInput     = "input"
	kOutput    = "output"
)

var (
	service   string
	tolang    string
	envfile   string
	proxy     string
	glossfile string
	input     string
	output    string
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
  cat input.txt | translate > output.txt
  translate -i input.txt -o output.txt`,
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
	cmd.Flags().StringVarP(&glossfile, KGlossFile, "g", "", "csv file for glossary")
	cmd.Flags().StringVarP(&input, kInput, "i", "", "input file, if set then stdin/pipe is ignored")
	cmd.Flags().StringVarP(&output, kOutput, "o", "", "output file, if set then stdout redirection is ignored")
	cmd.Flags().StringVarP(&proxy, kProxy, "p", "", "http or socks5 proxy,\n eg. http://127.0.0.1:7890 or socks5://127.0.0.1:7890")

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
	glossary, err := util.LoadGlossary(glossfile)
	if err != nil {
		return err
	}
	trans, err := translator.GetTranslator(service, proxy, glossary)
	if err != nil {
		return err
	}
	if c, ok := trans.(io.Closer); ok {
		defer c.Close()
	}

	reader, err := getInputReader(args)
	if err != nil {
		return err
	}
	if reader == nil { // 从终端交互读取要翻译的文本
		return translateInTerminal(trans)
	}
	if c, ok := reader.(io.Closer); ok {
		defer c.Close()
	}
	writer, err := getOutputWriter()
	if err != nil {
		return err
	}
	if c, ok := writer.(io.Closer); ok {
		defer c.Close()
	}

	o, ok := trans.(translator.TranslationObserver)
	if ok { // 收到分组响应后立即输出
		o.OnTranslated(func(translated []string) error {
			return util.WriteLines(writer, translated)
		})
	}
	texts, err := util.ReadLines(reader)
	if err != nil {
		return err
	}
	result, err := trans.Translate(texts, tolang)
	if err != nil {
		return err
	}
	if !ok { // 收到全部响应后再输出
		if err = util.WriteLines(writer, result); err != nil {
			return err
		}
	}

	return nil
}

func getInputReader(args []string) (io.Reader, error) {
	if input != "" { // 从-i读取要翻译的文本
		f, err := os.Open(input)
		if err != nil {
			return nil, fmt.Errorf("open input file: %w", err)
		}
		return f, nil
	}

	if !util.IsTerminal() { // 从os.Stdin读取要翻译的文本
		return os.Stdin, nil
	}

	if len(args) > 0 { // 翻译命令行参数
		return strings.NewReader(strings.Join(args, "\n")), nil
	}

	return nil, nil
}

func getOutputWriter() (io.Writer, error) {
	if output != "" { // 向-o写入翻译结果
		f, err := os.Create(output)
		if err != nil {
			return nil, fmt.Errorf("create output file: %w", err)
		}
		return f, nil
	}
	return os.Stdout, nil
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
