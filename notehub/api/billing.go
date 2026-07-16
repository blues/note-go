package api

// GetBillingAccountResponse v1
//
// The response object for getting a billing account.
type GetBillingAccountResponse struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	// "billing_admin", "billing_manager", "project_creator", or "billing_member"
	Role string `json:"role"`
}

type GetBillingAccountsResponse struct {
	BillingAccounts []GetBillingAccountResponse `json:"billing_accounts"`
}
