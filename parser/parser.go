package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type ProductInfo struct {
	Article   string `json:"article"`
	Price     string `json:"price"`
	OldPrice  string `json:"old_price"`
	CardPrice string `json:"card_price"`
}

// Start запускает парсер с указанным количеством страниц
func Start(maxPages int) {
	// Создание директории output, если она не существует
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		err := os.Mkdir("output", os.ModePerm)
		if err != nil {
			log.Fatalf("Ошибка при создании директории: %v", err)
		}
	}
	// Настройка контекста с увеличенным таймаутом
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	mainCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Увеличение общего таймаута
	mainCtx, cancel = context.WithTimeout(mainCtx, 30*time.Minute)
	defer cancel()

	var productInfos []ProductInfo
	baseURL := "https://www.ozon.ru"
	nextPage := "/brand/sokolov-136571558/?redirect_query=sokolov"
	visitedLinks := make(map[string]bool)
	productsProcessed := 0
	currentPage := 1

	for nextPage != "" && currentPage <= maxPages {
		fullURL := baseURL + nextPage
		var productLinks []string

		// Переход на страницу списка и извлечение ссылок на продукты
		err := chromedp.Run(mainCtx, chromedp.Tasks{
			chromedp.Navigate(fullURL),
			chromedp.Sleep(5 * time.Second), // Ожидание обхода защиты Cloudflare
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.iy7.iy8.tile-root a')).map(a => a.href)`, &productLinks),
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Найдено %d товаров на странице %s\n", len(productLinks), fullURL)

		// Обработка каждой ссылки на продукт
		for _, link := range productLinks {
			// Проверка и корректировка формата ссылки
			link = ensureAbsoluteURL(baseURL, link)

			// Проверка, была ли ссылка уже посещена
			if visitedLinks[link] {
				fmt.Printf("Ссылка на товар уже была посещена: %s\n", link)
				continue
			}
			visitedLinks[link] = true

			fmt.Printf("Обработка ссылки на товар: %s\n", link)

			var article, price, oldPrice, cardPrice string
			var itemUnavailable bool

			// Создание контекста с таймаутом для каждой страницы товара
			ctx, cancel := context.WithTimeout(mainCtx, 10*time.Second)
			err := chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Navigate(link),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),
				chromedp.Evaluate(`document.querySelector('.ul7') !== null`, &itemUnavailable),
				chromedp.WaitVisible(`.ga13-a2.tsBodyControl400Small`, chromedp.ByQuery),
				chromedp.Text(`.p1k`, &article, chromedp.ByQuery),
				chromedp.Text(`.lz7 .l5z.tsHeadline500Medium`, &price, chromedp.ByQuery),
				chromedp.Text(`.lz7 .z4l.z5l.z3l.tsBody400Small`, &oldPrice, chromedp.ByQuery),
				chromedp.Text(`.l8y .zl0.tsHeadline700XLarge`, &cardPrice, chromedp.ByQuery),
			})
			cancel()
			if err != nil {
				log.Println("Ошибка при получении деталей товара:", err)
				continue
			}

			// Проверка, доступен ли товар
			if itemUnavailable {
				fmt.Printf("Товар недоступен: %s\n", link)
				// Возврат на страницу списка
				err = chromedp.Run(mainCtx, chromedp.Tasks{
					chromedp.Navigate(fullURL),
					chromedp.Sleep(2 * time.Second), // Ожидание загрузки страницы
				})
				if err != nil {
					log.Println("Ошибка при возврате на страницу списка:", err)
					break
				}

				fmt.Printf("Возвращение на страницу списка: %s\n", fullURL)
				continue
			}

			productInfos = append(productInfos, ProductInfo{
				Article:   article,
				Price:     price,
				OldPrice:  oldPrice,
				CardPrice: cardPrice,
			})

			productsProcessed++
			fmt.Printf("Получены детали товара: Артикул=%s, Цена=%s, Старая цена=%s, Цена с картой=%s\n", article, price, oldPrice, cardPrice)

			// Сохранение промежуточных результатов каждые 50 товаров
			if productsProcessed%50 == 0 {
				saveToJSON("output/products_intermediate.json", productInfos)
			}

			// Возврат на страницу списка
			err = chromedp.Run(mainCtx, chromedp.Tasks{
				chromedp.Navigate(fullURL),
				chromedp.Sleep(2 * time.Second), // Ожидание загрузки страницы
			})
			if err != nil {
				log.Println("Ошибка при возврате на страницу списка:", err)
				break
			}

			fmt.Printf("Возвращение на страницу списка: %s\n", fullURL)
		}

		// Проверка, есть ли следующая страница
		var nextLink string
		err = chromedp.Run(mainCtx, chromedp.Tasks{
			chromedp.Evaluate(`document.querySelector('a[class="n5e b213-a0 b213-b6 b213-b1"]') ? document.querySelector('a[class="n5e b213-a0 b213-b6 b213-b1"]').href : ""`, &nextLink),
		})
		if err != nil {
			log.Fatal(err)
		}

		if nextLink != "" {
			// Проверка и корректировка формата следующей ссылки
			parsedNextLink, err := url.Parse(nextLink)
			if err != nil {
				log.Fatalf("Неверный URL следующей страницы: %v", err)
			}
			nextPage = parsedNextLink.RequestURI()
		} else {
			nextPage = ""
		}

		fmt.Printf("Ссылка на следующую страницу: %s\n", nextPage)
		currentPage++
	}

	// Сохранение окончательных результатов в JSON-файл
	saveToJSON("output/products.json", productInfos)

	fmt.Println("Информация о товарах сохранена в output/products.json")
}

// ensureAbsoluteURL проверяет, что URL абсолютный, разрешая относительные URL
func ensureAbsoluteURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Неверный базовый URL: %v", err)
	}
	ref, err := url.Parse(relativeURL)
	if err != nil {
		log.Fatalf("Неверный относительный URL: %v", err)
	}
	return u.ResolveReference(ref).String()
}

// saveToJSON сохраняет информацию о товарах в JSON-файл
func saveToJSON(filename string, productInfos []ProductInfo) {
	fmt.Printf("Сохранение результатов в %s...\n", filename) // Отладочное сообщение
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Ошибка при создании JSON-файла: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(productInfos)
	if err != nil {
		log.Fatalf("Ошибка при кодировании JSON: %v", err)
	}

	fmt.Printf("Результаты успешно сохранены в %s\n", filename)
}
