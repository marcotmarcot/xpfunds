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
	"math"
    "net/http"
	"regexp"
	"strconv"
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
	req.Header.Add("Cookie", cookie)
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
        if len(matches) < 2 {
			continue
        }
        if len(matches) > 2 {
            log.Fatal(len(matches))
        }
		entities = append(entities, matches[1])
    }
	return entities
}

func main() {
	nameRegexp := regexp.MustCompile("<h2 class=\"fleft\">(.*)</h2>")
	valuesRegexp := regexp.MustCompile("<td class=\"TD_blue\">(-?[0-9,]+)</td>")
	for _, fund := range entities(get("https://portal.xpi.com.br/pages/fundos/tabela-rentabilidades.aspx"), regexp.MustCompile("<a href=\"(/pages/fundos/fundos-investimentos.aspx\\?F=[0-9]+)\">")) {
		html := get("https://portal.xpi.com.br" + fund)
		name := entities(html, nameRegexp)[0]
		prod := 1.0
		values := entities(html, valuesRegexp)
		for _, sv := range values {
			v, err := strconv.ParseFloat(strings.Replace(sv, ",", ".", 1), 64)
			if err != nil {
				log.Fatal(err)
			}
			prod *= 1.0 + v / 100.0
		}
		fmt.Printf("%s\t%.2f%%\n", name, (math.Pow(prod, 1.0 / (float64(len(values)) / 12.0)) - 1.0) * 100.0)
	}
}
