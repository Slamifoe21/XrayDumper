package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	URL      string
	HWID     string
	OutFile  string
	Timeout  int
	Insecure bool
}

func main() {
	exitCode := mainImpl()

	if ownsConsole() {
		fmt.Println("Нажмите Enter для выхода...")
		fmt.Scanln()
	}

	os.Exit(exitCode)
}

func mainImpl() int {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nОшибка: %v", err)
		return 1
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "\nКритическая ошибка: %v", err)
		return 1
	}
	return 0
}

func run(cfg Config) error {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	if cfg.Insecure {
		tr, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return errors.New("не удалось получить дефолтный транспорт")
		}
		tr = tr.Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = tr
	}

	content, err := fetchSubscription(httpClient, cfg.URL, cfg.HWID)
	if err != nil {
		return fmt.Errorf("ошибка загрузки подписки: %w", err)
	}

	decodedBytes, err := decodeBase64(content)
	if err != nil {
		return fmt.Errorf("подписка не поддерживается: %w", err)
	}

	links, err := parseLinks(decodedBytes)
	if err != nil {
		return err
	}

	if err := saveLinks(cfg.OutFile, links); err != nil {
		return fmt.Errorf("ошибка сохранения файла: %w", err)
	}
	fmt.Printf("Успешно записано %d ссылок в файл %s!\n", len(links), cfg.OutFile)
	return nil
}

func parseFlags() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.URL, "url", "", "Ссылка на подписку")
	flag.StringVar(&cfg.HWID, "hwid", "0123456789abcdef", "Ваш HWID")
	flag.StringVar(&cfg.OutFile, "out", "links.txt", "Имя выходного файла")
	flag.IntVar(&cfg.Timeout, "timeout", 15, "Timeout для запроса")
	flag.BoolVar(
		&cfg.Insecure,
		"insecure",
		false,
		"Отключить проверку TLS-сертификата (небезопасно)",
	)

	flag.Parse()

	cfg.URL = strings.TrimSpace(cfg.URL)
	cfg.OutFile = strings.TrimSpace(cfg.OutFile)

	if cfg.Timeout <= 0 {
		return Config{}, errors.New("timeout должен быть больше 0")
	}

	if cfg.URL == "" {
		if ownsConsole() {
			fmt.Fprint(os.Stderr, "Введите ссылку на подписку: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return Config{}, fmt.Errorf("не удалось прочитать ввод пользователя: %w", err)
			}
			cfg.URL = strings.TrimSpace(input)
			if cfg.URL == "" {
				flag.Usage()
				return Config{}, errors.New("URL не может быть пустым")
			}
		} else {
			flag.Usage()
			return Config{}, errors.New("URL не может быть пустым")
		}
	}

	if !strings.HasPrefix(cfg.URL, "http://") && !strings.HasPrefix(cfg.URL, "https://") {
		return Config{}, errors.New("ссылка должна начинаться с http:// или https://")
	}

	if cfg.OutFile == "" {
		return Config{}, errors.New("имя выходного файла не может быть пустым")
	}

	return cfg, nil
}

func parseLinks(decodedBytes []byte) ([]string, error) {
	rawLines := strings.Split(strings.ReplaceAll(string(decodedBytes), "\r\n", "\n"), "\n")

	var links []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "://") {
			links = append(links, line)
		}
	}

	if len(links) == 0 {
		return nil, errors.New("не удалось найти ссылки в подписке")
	}
	return links, nil
}

func decodeBase64(raw string) ([]byte, error) {
	s := strings.Join(strings.Fields(raw), "")
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}

	return nil, errors.New("невалидный формат")
}

func fetchSubscription(httpClient *http.Client, url, hwid string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("создание запроса: %w", err)
	}

	req.Header.Set("User-Agent", "v2ray/1.0")
	req.Header.Set("X-App-Version", "1.0")
	req.Header.Set("X-HWID", hwid)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("отправка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("неожиданный статус: %s", resp.Status)
	}

	const maxBodySize = 10 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return "", fmt.Errorf("чтение ответа: %w", err)
	}

	if len(body) > maxBodySize {
		return "", fmt.Errorf("размер ответа превысил максимально допустимый лимит в %d байт", maxBodySize)
	}

	return string(body), nil
}

func saveLinks(outputFilename string, links []string) error {
	file, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("создание файла %q: %w", outputFilename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, link := range links {
		if _, err := fmt.Fprintln(writer, link); err != nil {
			return fmt.Errorf("запись ссылки: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("очистка буфера: %w", err)
	}

	return nil
}
