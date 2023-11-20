package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/tealeg/xlsx"
)

type SearchResult struct {
	Kod          string `json:"Kod"`
	SirketUnvanı string `json:"Şirket Unvanı"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		searchTerm := r.FormValue("search")
		searchTerm = strings.ToUpper(searchTerm)

		sirketlerExcelFilePath := "./exel/sirketler.xlsx"
		sirketlerFile, err := xlsx.OpenFile(sirketlerExcelFilePath)
		if err != nil {
			fmt.Print("şirketler dosyası açılırken hata", err)
			log.Fatal(err)
		}

		var results []SearchResult

		processRow := func(row *xlsx.Row) {
			var result SearchResult

			cellValue1 := row.Cells[0].String()
			cellValue2 := row.Cells[1].String()

			if strings.HasPrefix(cellValue1, searchTerm) || strings.HasPrefix(cellValue2, searchTerm) {
				fmt.Printf("Aranan kelime: \"%s\" bulunan: %s\n", searchTerm, cellValue1)

				result.Kod = cellValue1
				result.SirketUnvanı = cellValue2
				results = append(results, result)
			}
		}

		for _, sheet := range sirketlerFile.Sheets {
			for _, row := range sheet.Rows {
				processRow(row)
			}
		}

		jsonResponse, err := json.Marshal(results)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}
	renderSearchUsingTemplate(w)
}

func renderSearchUsingTemplate(w http.ResponseWriter) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func renderResultsUsingTemplate(w http.ResponseWriter, searchTerm string, results []string) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	data := struct {
		SearchTerm string
		Results    []string
	}{
		SearchTerm: searchTerm,
		Results:    results,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}
func companyPageHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	companyName := parts[len(parts)-1]
	unvan := r.URL.Query().Get("unvan")
	log.Printf("Tıklanan unvan: %s", unvan)

	excelFilePath := "./exel/bist-katilim-endekslerinde-yer-alan-paylarin-listesi.xlsx"
	xlFile, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		log.Printf("Excel dosyası açılırken hata: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	companyName = strings.ToUpper(companyName)
	var bist, bultenAdi string

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			if len(row.Cells) > 1 {
				payKodu := row.Cells[1].String()
				if payKodu == companyName {
					bist = "true"

					if len(row.Cells) > 2 {
						bultenAdi = row.Cells[2].String()
					}
					break
				}
			}
		}
		if bist != "" {
			break
		}
	}
	excelFilePath2 := "./exel/temettu.xlsx"
	xlFile2, err := xlsx.OpenFile(excelFilePath2)
	if err != nil {
		log.Printf("Excel dosyası açılırken hata: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	companyName = strings.ToUpper(companyName)
	var temettu string

	for _, sheet2 := range xlFile2.Sheets {
		for _, row2 := range sheet2.Rows {
			if len(row2.Cells) > 1 {
				payKodu2 := row2.Cells[0].String()
				if payKodu2 == companyName {
					temettu = "true"
					break
				}
			}
		}
		if temettu != "" {
			break
		}
	}
	renderCompanyPage(w, companyName, bist, bultenAdi, unvan, temettu)
}

func renderCompanyPage(w http.ResponseWriter, companyName, bist, bultenAdi, unvan, temmettu string) {
	tmpl, err := template.ParseFiles("company.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		CompanyName string
		BIST        string
		BultenAdi   string
		Unvan       string
		Temettu     string
	}{
		CompanyName: companyName,
		BIST:        bist,
		BultenAdi:   bultenAdi,
		Unvan:       unvan,
		Temettu:     temmettu,
	}
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", searchHandler)
	http.HandleFunc("/company/", companyPageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server is listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
