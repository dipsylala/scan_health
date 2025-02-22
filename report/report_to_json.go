package report

import (
	"encoding/json"
	"fmt"
	"github.com/antfie/scan_health/v2/utils"
	"os"
)

func (r *Report) renderToJson(filePath string) {
	formattedJson, err := json.MarshalIndent(r, "", "    ")

	if err != nil {
		utils.ErrorAndExit("Could not render to JSON", err)
	}

	if filePath != "" {
		if err := os.WriteFile(filePath, formattedJson, 0666); err != nil {
			utils.ErrorAndExit("Could not save JSON file", err)
		}
		return
	}

	fmt.Println(string(formattedJson))
}
