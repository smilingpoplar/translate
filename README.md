将文本翻译成中文，支持 google 翻译、openai 翻译

## 用法

```sh
go install github.com/smilingpoplar/translate/cmd/translate@latest
```

```sh
translate "hello world"

cat input.txt | translate > output.txt
```

### 使用 Docker

```sh
docker build -t translate .
```

```sh
docker run --rm translate "hello world"

cat input.txt | docker run --rm -i translate > output.txt
```
