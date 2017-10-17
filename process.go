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
		fmt.Printf("\t%d", cot+liq)
		sum := 0.0
		neg := 0
		var values []float64
		for i := 4; i < len(fields); i++ {
			v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
			if err != nil {
				log.Fatal(err)
			}
			v = v / 100.0
			values = append(values, v)
			sum += v
			if v < 0 {
				neg++
			}
		}
		prod := 1.0
		mean := sum / float64(len(values))
		total := 0.0
		gd := 1.0
		gds := 0
		for i := len(values) - 1; i >= 0; i-- {
			prod *= 1.0 + values[i]
			total += math.Pow(values[i]-mean, 2)
			sprod := 1.0
			for j := i; j >= 0; j-- {
				sprod *= 1.0 + values[j]
				if sprod < gd {
					gd = sprod
					gds = i - j + 1
				}
			}
		}
		fmt.Printf("\t%d\t%s\t%s%%\t%s%%\t%d\t%s%%\n", len(values), formatFloat(100.0*math.Sqrt(total/float64(len(values)))), formatFloat(100.0*float64(neg)/float64(len(values))), formatFloat((gd-1.0)*100.0), gds, formatFloat((math.Pow(prod, 1.0/(float64(len(values))/12.0))-1.0)*100.0))
	}
}

func formatFloat(f float64) string {
	return strings.Replace(strconv.FormatFloat(f, 'f', 2, 64), ".", ",", 1)
}

type fund struct {
	name string
	min float64
	days int64
	raw []float64
}
