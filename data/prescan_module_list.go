package data

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/antfie/scan_health/v2/report"
	"github.com/antfie/scan_health/v2/utils"
)

type prescanModuleList struct {
	XMLName xml.Name        `xml:"prescanresults"`
	Modules []prescanModule `xml:"module"`
}

type prescanModule struct {
	XMLName        xml.Name             `xml:"module"`
	Id             int                  `xml:"id,attr"`
	Name           string               `xml:"name,attr"`
	Status         string               `xml:"status,attr"`
	Platform       string               `xml:"platform,attr"`
	Size           string               `xml:"size,attr"`
	MD5            string               `xml:"checksum,attr"`
	HasFatalErrors bool                 `xml:"has_fatal_errors,attr"`
	IsDependency   bool                 `xml:"is_dependency,attr"`
	Files          []prescanFileIssue   `xml:"file_issue"`
	Issues         []prescanModuleIssue `xml:"issue"`
	SizeBytes      int
}

type prescanFileIssue struct {
	XMLName xml.Name `xml:"file_issue"`
	Name    string   `xml:"filename,attr"`
	Details string   `xml:"details,attr"`
}

type prescanModuleIssue struct {
	XMLName xml.Name `xml:"issue"`
	Details string   `xml:"details,attr"`
}

func populateModuleInstances(r *report.Report, moduleList prescanModuleList) {
	for _, module := range moduleList.Modules {
		var issues []string

		for _, issue := range module.Issues {
			formattedIssueDetails := html.UnescapeString(issue.Details)

			if !utils.IsStringInStringArray(formattedIssueDetails, issues) {
				issues = append(issues, formattedIssueDetails)
			}

		}

		formattedStatus := html.UnescapeString(module.Status)

		if formattedStatus != "OK" {
			statusParts := strings.Split(formattedStatus, ",")

			for _, statusPart := range statusParts {
				formattedStatusPart := strings.TrimSpace(statusPart)

				if !utils.IsStringInStringArray(formattedStatusPart, issues) {
					issues = append(issues, formattedStatusPart)
				}
			}
		}

		r.AddModuleInstance(
			html.UnescapeString(module.Name),
			report.ModuleInstance{
				Id:             module.Id,
				Status:         html.UnescapeString(module.Status),
				Platform:       html.UnescapeString(module.Platform),
				Size:           html.UnescapeString(module.Size),
				MD5:            module.MD5,
				HasFatalErrors: module.HasFatalErrors,
				IsDependency:   module.IsDependency,
				Issues:         issues,
				Source:         "prescan_module_list",
				SizeBytes:      calculateModuleSize(module.Size),
			},
		)
	}
}

func calculateModuleSize(size string) int {
	var totalModuleSize = 0
	totalModuleSize += convertSize(size, "GB", 1e+9)
	totalModuleSize += convertSize(size, "MB", 1e+6)
	totalModuleSize += convertSize(size, "KB", 1000)
	return totalModuleSize
}

func convertSize(size, measurement string, multiplier int) int {
	if !strings.HasSuffix(size, measurement) {
		return 0
	}

	formattedSize := strings.TrimSuffix(size, measurement)
	sizeInt, err := strconv.Atoi(formattedSize)

	if err != nil {
		panic(err)
	}

	return sizeInt * multiplier
}

func (api API) populatePrescanModuleList(r *report.Report) {

	var url = fmt.Sprintf("/api/5.0/getprescanresults.do?app_id=%d&build_id=%d", r.Scan.ApplicationId, r.Scan.BuildId)
	response := api.makeApiRequest(url, http.MethodGet)

	moduleList := prescanModuleList{}

	err := xml.Unmarshal(response, &moduleList)

	if err != nil {
		// We may not have this data for older scans, but it is no reason to panic
		if r.IsReportForOtherScan {
			return
		}

		utils.ErrorAndExit("Could not parse getprescanresults.do API response", err)
	}

	populateModuleInstances(r, moduleList)
}
