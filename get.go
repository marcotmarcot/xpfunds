// This program accesses XP website and calculates the annualized return rate of
// each fund during its whole existence. You need to inform the Cookie that your
// session is using. To do so, use the inspection function of your web browser
// and see the Cookie header that is being sent after you login to XP.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	cookie = flag.String("cookie", "", "The cookie to be used to login to XP")
	client = &http.Client{}
)

func get(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Cookie", *cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return string(b)
}

func entities(html string, r *regexp.Regexp) []string {
	var entities []string
	for _, line := range strings.Split(html, "\n") {
		matches := r.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		if len(matches) != 2 {
			log.Fatal(len(matches))
		}
		entities = append(entities, matches[1])
	}
	return entities
}

func nextLine(fund string, html string, pr *regexp.Regexp, r *regexp.Regexp) string {
	var i int
	var line string
	lines := strings.Split(html, "\n")
	for i, line = range lines {
		matches := pr.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		break
	}
	i++
	if i >= len(lines) {
		log.Fatalf("Could not find %v on %v", pr, fund)
	}
	return r.FindStringSubmatch(lines[i])[1]
}

func main() {
	flag.Parse()
	nameR := regexp.MustCompile("<h2 class=\"fleft\">(.*)</h2>")
	openR := regexp.MustCompile("(Quero aplicar agora)")
	qualR := regexp.MustCompile("(qualificado)")
	minPR := regexp.MustCompile("Aplicação Inicial Mínima")
	minR := regexp.MustCompile("<div class=\"TB_EspacComp TB_EspacCompValor\">R\\$ ([0-9,.]*)</div>")
	cotPR := regexp.MustCompile("Resgate - Cotização")
	cotR := regexp.MustCompile("<div class=\"TB_EspacComp TB_EspacCompValor\">D\\+([0-9]*)")
	liqPR := regexp.MustCompile("Resgate - Liquidação Financeira")
	liqR := regexp.MustCompile("<div class=\"TB_EspacComp TB_EspacCompValor\">D\\+([0-9]*)")
	valuesR := regexp.MustCompile("<td class=\"TD_blue\">(-?[0-9,]+)</td>")
	for _, fund := range entities(get("https://portal.xpi.com.br/pages/fundos/tabela-rentabilidades.aspx"), regexp.MustCompile("<a href=\"(/pages/fundos/fundos-investimentos.aspx\\?F=[0-9]+)\">")) {
		html := get("https://portal.xpi.com.br" + fund)
		if entities(html, openR) == nil {
			continue
		}
		if entities(html, qualR) != nil {
			log.Println(fund)
			continue
		}
		fmt.Printf("%s\t%s\t%s\t%s", entities(html, nameR)[0], nextLine(fund, html, minPR, minR), nextLine(fund, html, cotPR, cotR), nextLine(fund, html, liqPR, liqR))
		for _, sv := range entities(html, valuesR) {
			fmt.Printf("\t%s", sv)
		}
		fmt.Printf("\n")
	}
}
