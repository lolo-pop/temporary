package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func zfill(str string, width int) string {
	for len(str) < width {
		str = "0" + str
	}

	return str
}
func main() {
	// 打开CSV文件
	file, err := os.Open("/home/rongch05/openfaas/faas-scaling/profiling.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 创建CSV Reader对象
	reader := csv.NewReader(file)

	// 读取CSV文件中的所有记录
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// 将CSV记录解析为字典
	results := make(map[string][]float64)
	for i, record := range records {
		if i == 0 {
			continue
		} else {
			configuration := record[0] + zfill(record[1], 4) + zfill(record[2], 2) + record[3]
			acc, err := strconv.ParseFloat(record[4], 64)
			if err != nil {
				panic(err)
			}
			lat1, err := strconv.ParseFloat(record[5], 64)
			if err != nil {
				panic(err)
			}
			lat2, err := strconv.ParseFloat(record[6], 64)
			if err != nil {
				panic(err)
			}
			results[configuration] = []float64{acc, lat1, lat2}
		}
	}

	// 打印字典
	fmt.Println(results["11024041"])
}
