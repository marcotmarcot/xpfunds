// This program accesses XP website and calculates the annualized return rate of
// each fund during its whole existence. You need to inform the Cookie that your
// session is using. To do so, use the inspection function of your web browser
// and see the Cookie header that is being sent after you login to XP.
package main

import (
	"bufio"
    "fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func formatFloat(f float64) string {
	return strings.Replace(strconv.FormatFloat(f, 'f', 2, 64), ".", ",", 1)
}

func main() {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("Nome\tMínimo\tDias para resgate\tIdade em meses\tDesvio padrão\tMeses negativos\tMaior queda\tPeríodo da maior queda em meses\tRentabilidade anualizada\n")
	for true {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		fields := strings.Split(strings.Trim(line, "\n"), "\t")
		if fields == nil {
			break
		}
		fmt.Printf("%s", fields[0])
		fmt.Printf("\t%s", fields[1])
		cot, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Fatal(err)
		}
		liq, err := strconv.Atoi(fields[3])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\t%d", cot + liq)
		prod := 1.0
		sum := 0.0
		neg := 0
		var values []float64
		for i := 4; i < len(fields); i++ {
			v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
			if err != nil {
				log.Fatal(err)
			}
			v /= 100.0
			prod *= 1.0 + v
			sum += v
			values = append(values, v)
			if (v < 0) {
				neg++
			}
		}
		mean := sum / float64(len(values))
		total := 0.0
		for _, v := range values {
			total += math.Pow(v-mean, 2)
		}
		gd := 1.0
		gds := 0
		for i := range values {
			prod := 1.0
			for j := i; j < len(values); j++ {
				prod *= 1.0 + values[j]
				if prod < gd {
					gd = prod
					gds = j - i + 1
				}
			}
		}
		fmt.Printf("\t%d\t%s\t%s%%\t%s%%\t%d\t%s%%\n", len(values), formatFloat(100.0 * math.Sqrt(total / float64(len(values)))), formatFloat(100.0 * float64(neg) / float64(len(values))), formatFloat((gd - 1.0) * 100.0), gds, formatFloat((math.Pow(prod, 1.0 / (float64(len(values)) / 12.0)) - 1.0) * 100.0))
	}
}
