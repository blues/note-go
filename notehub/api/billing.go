package api

// GetBillingAccountResponse v1
//
// The response object for getting a billing account.
type GetBillingAccountResponse struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	// "billing_admin", "billing_manager", or "project_creator"
	Role string `json:"role"`
}

type GetBillingAccountsResponse struct {
	BillingAccounts []GetBillingAccountResponse `json:"billing_accounts"`
}
