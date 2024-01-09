package usecase

import (
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

type generateLetterUsecase struct {
	Sheets *sheets.Service
	Docs   *docs.Service
	Drive  *drive.Service
}

type GenerateLetterResponse struct {
	Email     string `json:"email"`
	URL       string `json:"url"`
	IsSuccess bool   `json:"is_success"`
}

type GenerateLetterRequest struct {
	Email []string `json:"email"`
}

type SheetsData struct {
	EmployeeID           string
	Name                 string
	Email                string
	Departement          string
	BaseCurrency         string
	BasePay              string
	ChangeBasePay        string
	RaiseEffectiveDate   string
	StockQuantity        string
	VestingDate          string
	BonusStructureChange string
	BonusEffectiveDate   string
}

func New(sheets *sheets.Service, docs *docs.Service, drive *drive.Service) GenerateLetterUsecase {
	return generateLetterUsecase{
		Sheets: sheets,
		Docs:   docs,
		Drive:  drive,
	}
}

func (gl generateLetterUsecase) GenerateLetter(form GenerateLetterRequest) (res []GenerateLetterResponse, err error) {
	spreadSheetId := "1PKSLn3MAo08O25Hj73OLtP-mP80Z4CQfsa-zJA5PTTE"
	readRange := "Sheet1"
	sheetData, err := gl.mappingDataSheets(spreadSheetId, readRange)
	if err != nil {
		return res, nil
	}

	fileID := "1mh_RM0Xu-M4N58sCZcdCSVqQOl-k9GeXFKkyhnAcCPM"
	for _, email := range form.Email {
		tmp := GenerateLetterResponse{}

		if val, ok := sheetData[email]; ok {
			tmp.IsSuccess = true
			tmp.Email = email

			documentID, err := gl.copyDocumentTemplate(fileID, fmt.Sprintf("%s %s", val.Name, time.Now().Format("2006-01-02")))
			if err != nil {
				return res, nil
			}

			err = gl.batchUpdateDocument(documentID, val)
			if err != nil {
				return res, nil
			}

			tmp.URL = fmt.Sprintf("https://docs.google.com/document/d/%s", documentID)
		}

		res = append(res, tmp)
	}

	return res, nil
}

func (gl generateLetterUsecase) batchUpdateDocument(documentID string, form SheetsData) error {
	gl.Docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Preferred Name}}",
					},
					ReplaceText: form.Name,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{#}}",
					},
					ReplaceText: form.EmployeeID,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Base Currency}}",
					},
					ReplaceText: form.BaseCurrency,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Change Base Pay Request}}",
					},
					ReplaceText: form.ChangeBasePay,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Raise Effective Date}}",
					},
					ReplaceText: form.RaiseEffectiveDate,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Stock Quantity}}",
					},
					ReplaceText: form.StockQuantity,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Vesting Date}}",
					},
					ReplaceText: form.VestingDate,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Bonus Structure Change}}",
					},
					ReplaceText: form.BonusStructureChange,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{Bonus Effective Date}}",
					},
					ReplaceText: form.BonusEffectiveDate,
				},
			},
			{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						Text: "{{date}}",
					},
					ReplaceText: time.Now().Format("2006-01-02"),
				},
			},
		},
	}).Do()

	return nil
}

func (gl generateLetterUsecase) copyDocumentTemplate(fileID, fileName string) (documentID string, err error) {
	copyRes, err := gl.Drive.Files.Copy(fileID, &drive.File{
		Name: fileName,
	}).Do()
	if err != nil {
		log.Fatal(err)
		return documentID, err
	}

	documentID = copyRes.Id

	return documentID, nil
}

func (gl generateLetterUsecase) mappingDataSheets(spreadSheetId, readRange string) (map[string]SheetsData, error) {
	sheetDataByEmail := make(map[string]SheetsData)

	resp, err := gl.Sheets.Spreadsheets.Values.Get(spreadSheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return sheetDataByEmail, err
	}

	if len(resp.Values) == 0 {
		return sheetDataByEmail, errors.New("no data found")
	} else {
		for i, row := range resp.Values {
			if i > 0 {
				sheetData := SheetsData{}
				sheetData.EmployeeID = fmt.Sprint(row[0])
				sheetData.Name = fmt.Sprint(row[1])
				sheetData.Email = fmt.Sprint(row[2])
				sheetData.Departement = fmt.Sprint(row[3])
				sheetData.BaseCurrency = fmt.Sprint(row[4])
				sheetData.BasePay = fmt.Sprint(row[5])
				sheetData.ChangeBasePay = fmt.Sprint(row[6])
				sheetData.RaiseEffectiveDate = fmt.Sprint(row[7])
				sheetData.StockQuantity = fmt.Sprint(row[8])
				sheetData.VestingDate = fmt.Sprint(row[9])
				sheetData.BonusStructureChange = fmt.Sprint(row[10])
				sheetData.BonusEffectiveDate = fmt.Sprint(row[11])

				sheetDataByEmail[fmt.Sprint(row[2])] = sheetData
			}
		}
	}

	return sheetDataByEmail, nil
}

type GenerateLetterUsecase interface {
	GenerateLetter(form GenerateLetterRequest) (res []GenerateLetterResponse, err error)
}
