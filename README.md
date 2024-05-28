# Ozon Product Scraper

Этот проект представляет собой парсер для сайта Ozon, который собирает информацию о продуктах с определенного бренда и сохраняет их в JSON-файл. Парсер использует библиотеку `chromedp` для управления браузером Google Chrome для обхода защиты **CloudFlare**.

## Установка

### Предварительные требования

- Установите [Go](https://golang.org/doc/install)
- Установите [Google Chrome](https://www.google.com/intl/ru_ru/chrome/)

### Клонирование репозитория

```bash
git clone https://github.com/thedeemoonn/ozon-product-scraper.git
cd ozon-product-scraper
```
Установка зависимостей
```
go mod tidy
```

## Запуск

### Запуск парсера
```
go run main.go 
```

## Использование

Парсер будет собирать информацию о продуктах и сохранять результаты в JSON-файлы в директорию output:

	•	output/products_intermediate.json — промежуточные результаты, сохраняемые каждые 50 продуктов.
	•	output/products.json — окончательные результаты после завершения парсинга.

### Поддержка

Если у вас есть вопросы или проблемы, пожалуйста, создайте [issue](https://github.com/thedeemoonn/parser/issues).

### Вклад

Приветствуется вклад в проект! Пожалуйста, создайте ветку для вашей функции или исправления и отправьте Pull Request.

	1.	Форкните репозиторий
	2.	Создайте новую ветку (git checkout -b feature/your-feature)
	3.	Внесите изменения и закоммитьте их (git commit -am 'Add new feature')
	4.	Отправьте изменения в вашу ветку (git push origin feature/your-feature)
	5.	Создайте новый Pull Request