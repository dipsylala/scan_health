package checks

import (
	"fmt"
	"github.com/antfie/scan_health/v2/report"
	"github.com/antfie/scan_health/v2/utils"
	"strings"
)

// Test cases
// Negative:
// https://analysiscenter.veracode.com/auth/index.jsp#ReviewResultsStaticFlaws:29692:1206933:27113269:27084165:27099815:::3763747 - has JS files module and is selected
//
// Positive:
//

func unselectedJavaScriptModules(r *report.Report) {
	var foundModules []string

	for _, module := range r.Modules {
		// Only applicable for JavaScript modules
		if !module.IsJavaScriptModule() {
			continue
		}

		// We don't care for node modules
		if module.IsNodeModule() {
			continue
		}

		// We don't care for modules from .map files
		if strings.Contains(strings.ToLower(module.Name), "extracted from .map file") {
			continue
		}

		if !module.HasFatalErrors() && !module.IsIgnored && !module.IsSelected() && !module.IsThirdParty {
			if !utils.IsStringInStringArray(module.Name, foundModules) {
				foundModules = append(foundModules, module.Name)
			}
		}
	}

	if len(foundModules) == 0 {
		return
	}

	message := fmt.Sprintf("A JavaScript module was not selected for analysis: \"%s\".", foundModules[0])

	if len(foundModules) > 1 {
		message = fmt.Sprintf("%d JavaScript modules were not selected for analysis: %s.", len(foundModules), utils.Top5StringList(foundModules))
	}

	r.ReportModuleIssue(message, report.IssueSeverityMedium, foundModules)
	r.MakeRecommendation("Veracode extracts JavaScript modules from the upload. Consider selecting the appropriate \"JS files within ...\" modules for analysis in order to cover the JavaScript risk from these components.")
	r.MakeRecommendation("Under-selection of first party modules affects results quality. Ensure the correct entry points have been selected as recommended and refer to this article: https://community.veracode.com/s/article/What-are-Modules-and-how-do-my-results-change-based-on-what-I-select.")
}
