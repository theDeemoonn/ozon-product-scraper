package main

import (
	"awesomeProject1/parser"
	"awesomeProject1/setup"
)

func main() {
	// Проверка наличия Google Chrome и ввод количества страниц
	maxPages := setup.Initialize()

	// Запуск парсера с указанным количеством страниц
	parser.Start(maxPages)
}
