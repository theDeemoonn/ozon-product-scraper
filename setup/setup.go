package setup

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Initialize проверяет наличие Google Chrome и запрашивает количество страниц для парсинга
func Initialize() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Установлен ли у вас Google Chrome? (yes/no): ")
	chromeInstalled, _ := reader.ReadString('\n')
	chromeInstalled = strings.TrimSpace(chromeInstalled)

	if strings.ToLower(chromeInstalled) != "yes" {
		fmt.Println("Пожалуйста, скачайте и установите Google Chrome с https://www.google.com/intl/ru_ru/chrome/")
		os.Exit(1)
	}

	fmt.Print("Введите количество страниц для парсинга: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	maxPages, err := strconv.Atoi(input)
	if err != nil {
		log.Fatalf("Неверный ввод для количества страниц: %v", err)
	}

	return maxPages
}
