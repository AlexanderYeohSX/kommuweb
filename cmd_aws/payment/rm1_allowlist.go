package main

import (
	"strings"
)

var devTestEmails = map[string]struct{}{
	"alexanderyeoh@kommu.ai": {},
	"keanwei@kommu.ai":       {},
}

const devTestAmountRM = "1.00"

func isDevTestEmail(email string) bool {
	e := strings.ToLower(strings.TrimSpace(email))
	_, ok := devTestEmails[e]
	return ok
}

// applyDevTestOrderOverride forces RM1 charge for allowlisted developers.
func applyDevTestOrderOverride(email string, amount string) (newAmount string, isTest bool) {
	if !isDevTestEmail(email) {
		return amount, false
	}
	return devTestAmountRM, true
}

// applyDevTestSubscriptionOverride forces RM1 deposit + optional test plan.
func applyDevTestSubscriptionOverride(email, deposit, planID string) (newDeposit, newPlan string, isTest bool) {
	if !isDevTestEmail(email) {
		return deposit, planID, false
	}
	testPlan := strings.TrimSpace(loadAppConfig().CurlecDevTestPlanID)
	if testPlan == "" {
		testPlan = planID
	}
	return devTestAmountRM, testPlan, true
}
