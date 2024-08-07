package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/smilingpoplar/translate/translator"
	"github.com/smilingpoplar/translate/translator/google"
	"github.com/smilingpoplar/translate/translator/openai"
	"github.com/smilingpoplar/translate/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	kEngine = "engine"
	kGoogle = "google"
	kOpenAI = "openai"
	kTolang = "tolang"
	kProxy  = "proxy"

	kOaiAPIKey  = "openai.api-key"
	kOaiBaseURL = "openai.base-url"
	kOaiModel   = "openai.model"
	kOaiPrompt  = "openai.prompt"
)

var (
	engine string
	tolang string
	proxy  string

	oaiAPIKey  string
	oaiBaseURL string
	oaiModel   string
	oaiPrompt  string
)

func main() {
	cmd := initCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	if err := initConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd := &cobra.Command{
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

	cmd.Flags().StringVarP(&engine, kEngine, "e", viper.GetString(kEngine), "translate engine, eg. google, openai")
	cmd.Flags().StringVarP(&tolang, kTolang, "t", viper.GetString(kTolang), "target language")
	cmd.Flags().StringVarP(&proxy, kProxy, "p", viper.GetString(kProxy), "http or socks5 proxy,\n eg. http://127.0.0.1:7890 or socks5://127.0.0.1:7890")
	cmd.Flags().StringVarP(&oaiAPIKey, kOaiAPIKey, "", viper.GetString(kOaiAPIKey), "required: openai, use any string if your openai compatible\n service (such as ollama) needs no key\n")
	cmd.Flags().StringVarP(&oaiBaseURL, kOaiBaseURL, "", viper.GetString(kOaiBaseURL), "optional: openai")
	cmd.Flags().StringVarP(&oaiModel, kOaiModel, "", viper.GetString(kOaiModel), "optional: openai")
	cmd.Flags().StringVarP(&oaiPrompt, kOaiPrompt, "", viper.GetString(kOaiPrompt), "optional: openai")
	cmd.MarkFlagsMutuallyExclusive(kOaiPrompt, kTolang)

	viper.BindPFlags(cmd.Flags())

	return cmd
}

func initConfig() error {
	viper.SetDefault(kEngine, kGoogle)
	viper.SetDefault(kTolang, "zh-CN")
	viper.SetDefault(kOaiBaseURL, openai.BaseURL)
	viper.SetDefault(kOaiModel, openai.DefaultModel)

	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("error find homedir, %w", err)
	}
	configDir := home + "/.config/translate/"
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error mkdir %s, %w", configDir, err)
	}
	if err := viper.ReadInConfig(); err != nil {
		// 第一次运行，没有配置文件
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("error write config, %w", err)
		}
	}
	viper.AutomaticEnv()
	// 将viper.Get(key)的key中'.'和'-'替换为'_'
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	return nil
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
	if engine == kGoogle {
		trans, err = google.New(google.WithProxy(proxy))
	} else if engine == kOpenAI {
		trans, err = getTranslatorOpenAI()
	} else {
		err = fmt.Errorf("unsupported engine: %s", engine)
	}
	return trans, err
}

func getTranslatorOpenAI() (translator.Translator, error) {
	API_KEY, BASE_URL := "OPENAI_API_KEY", "OPENAI_BASE_URL"
	if oaiAPIKey == "" {
		msg := fmt.Sprintf("%s is not set, set it with `export %s=YOUR_API_KEY`", API_KEY, API_KEY)
		msg += fmt.Sprintf("\n%s is %q, set it with `export %s=YOUR_BASE_URL`", BASE_URL, oaiBaseURL, BASE_URL)
		return nil, fmt.Errorf(msg)
	}
	return openai.New(oaiAPIKey,
		openai.WithBaseURL(oaiBaseURL),
		openai.WithModel(oaiModel),
		openai.WithPrompt(oaiPrompt),
		openai.WithProxy(proxy))
}
