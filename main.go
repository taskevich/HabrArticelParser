package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	BASEURL = "https://habr.com/ru/feed/"
)

type Item struct {
	ArticelTitle      string
	ArticleUrl        string
	ArticelComplexity string
	ArticleViews      string
	ArticelTimeRead   string
	ArticelImage      string
}

func getPagesCount(c *colly.Collector) int {
	c.Visit(BASEURL)
	ch := make(chan int, 1)
	c.OnHTML(".tm-pagination__pages", func(e *colly.HTMLElement) {
		pages := e.ChildTexts("a[class=tm-pagination__page]")
		pagesNum, err := strconv.Atoi(pages[len(pages)-1])
		if err != nil {
			ch <- 0
		} else {
			ch <- pagesNum
		}
	})
	c.Wait()
	c.OnHTMLDetach(".tm-pagination__pages")
	return <-ch
}

func writeFile(articles []Item) (bool, error) {
	b, err := json.Marshal(articles)
	if err != nil {
		return false, err
	}

	err = os.WriteFile("articles.txt", b, 0755)

	if err != nil {
		return false, err
	}

	return true, nil
}

func main() {
	articles := make([]Item, 1)

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains("habr.com"),
	)

	c.OnHTML(".tm-articles-list__item", func(e *colly.HTMLElement) {
		item := Item{}
		item.ArticelTitle = e.ChildText("a[class=tm-title__link]")
		if len(item.ArticelTitle) > 0 {
			item.ArticleUrl = e.ChildAttr("a[class=tm-title__link]", "href")
			item.ArticelComplexity = e.ChildText("span[class=tm-article-complexity__label]")
			item.ArticleViews = e.ChildText("span[class=tm-icon-counter__value]")
			item.ArticelTimeRead = e.ChildText("span[classtm-article-reading-time__label]")
			item.ArticelImage = e.ChildAttr("img[class=tm-article-snippet__lead-image]", "src")
			articles = append(articles, item)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	pagesNum := getPagesCount(c)

	for i := 1; i <= pagesNum; i++ {
		c.Visit(fmt.Sprintf("https://habr.com/ru/feed/page%d/", i))
	}

	c.Wait()

	ok, err := writeFile(articles)

	if !ok {
		panic(err)
	}
}
