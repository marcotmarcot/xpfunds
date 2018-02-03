package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"log"
	"net/http"
	"strings"
)

var (
	cookie = flag.String("cookie", "", "The cookie to be used to login to XP")
	client = &http.Client{}
)

func main() {
	flag.Parse()
	doc := get("https://portal.xpi.com.br/pages/fundos/tabela-rentabilidades.aspx")
	doc.Find("a[href^=\"/pages/fundos/fundos-investimentos.aspx?F=\"]").Each(func(index int, fund *goquery.Selection) {
		href, ok := fund.Attr("href")
		if !ok {
			log.Fatal("no href")
		}
		processFund(href)
	})
}

func get(url string) *goquery.Document {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Cookie", *cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func processFund(url string) {
	doc := get("https://portal.xpi.com.br" + url)
	if doc.Find("input[value=\"Quero aplicar agora\"]").Length() != 1 || doc.FindMatcher(suffixMatcher{"ualificados"}).Length() == 1 {
		return
	}
	name := doc.Find("h2.fleft").Text()
	min := doc.FindMatcher(suffixMatcher{"Aplicação Inicial Mínima"}).Next().Text()[3:]
	cot := strings.Split(strings.Split(doc.FindMatcher(suffixMatcher{"Resgate - Cotização"}).Next().Text(), "(")[0], " ")[0][2:]
	prefixLen := 2
	liq := doc.FindMatcher(suffixMatcher{"Resgate - Liquidação Financeira"}).Next().Text()
	if strings.HasPrefix(liq, "D+ ") {
		prefixLen = 3
	}
	liq = strings.Split(strings.Split(liq[prefixLen:], "(")[0], " ")[0]
	var allProfs []string
	doc.FindMatcher(fundYearMatcher{}).Each(func(index int, year *goquery.Selection) {
		var yearProfs []string
		year.Children().Each(func(index int, month *goquery.Selection) {
			prof := month.Text()
			if prof == "Fundo" || prof == "-" {
				return
			}
			yearProfs = append([]string{prof}, yearProfs...)
		})
		allProfs = append(allProfs, yearProfs...)
	})
	fmt.Println(strings.Join(append([]string{name, min, cot, liq}, allProfs...), "\t"))
}

type fundYearMatcher struct {
}

func (m fundYearMatcher) Match(n *html.Node) bool {
	if n.DataAtom != atom.Tr || len(n.Attr) != 0 || n.FirstChild == nil {
		return false
	}
	td := n.FirstChild.NextSibling
	if td.DataAtom != atom.Td || len(td.Attr) != 1 {
		return false
	}
	class := td.Attr[0]
	return class.Namespace == "" && class.Key == "class" && class.Val == "TD_colDesc TD_blue" && td.FirstChild != nil && td.FirstChild.Data == "Fundo"
}

func (m fundYearMatcher) MatchAll(n *html.Node) []*html.Node {
	return matchAll(m, n)
}

func (m fundYearMatcher) Filter(ns []*html.Node) []*html.Node {
	return filter(m, ns)
}

type suffixMatcher struct {
	suffix string
}

func (m suffixMatcher) Match(n *html.Node) bool {
	return n.FirstChild != nil && strings.HasSuffix(n.FirstChild.Data, m.suffix)
}

func (m suffixMatcher) MatchAll(n *html.Node) []*html.Node {
	return matchAll(m, n)
}

func (m suffixMatcher) Filter(ns []*html.Node) []*html.Node {
	return filter(m, ns)
}

func matchAll(m goquery.Matcher, n *html.Node) []*html.Node {
	var matches []*html.Node
	if m.Match(n) {
		matches = append(matches, n)
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		matches = append(matches, matchAll(m, child)...)
	}
	return matches
}

func filter(m goquery.Matcher, ns []*html.Node) []*html.Node {
	var matches []*html.Node
	for _, n := range ns {
		if m.Match(n) {
			matches = append(matches, n)
		}
	}
	return matches
}
