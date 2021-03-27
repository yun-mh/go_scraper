package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

func main() {
	var jobsTotal []extractedJob
	totalPages := getPages()

	for i := 0; i < totalPages; i++ {
		jobs := getPage(i)
		jobsTotal = append(jobsTotal, jobs...)
	}

	writeJobs(jobsTotal)
	fmt.Println(strconv.Itoa(len(jobsTotal)) + "件の抽出完了！")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"リンク", "タイトル", "位置", "年収", "備考"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/jobs?q=python&vjk=" + job.id, job.title, job.location, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

}

func getPage(page int) []extractedJob {
	var jobs []extractedJob

	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("リクエスト中：", pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")

	searchCards.Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobs = append(jobs, job)
	})
	return jobs
}

func extractJob(card *goquery.Selection) extractedJob {
	id, _ := card.Attr("data-jk")
	title := cleanString(card.Find(".title>a").Text())
	location := cleanString(card.Find(".sjcl>span").Text())
	salary := cleanString(card.Find(".salarySnippet>span>span").Text())
	summary := cleanString(card.Find(".summary").Text())

	return extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

func getPages() int {
	pages := 0

	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("リクエストが失敗しました。ステータスコード：", res.StatusCode)
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
