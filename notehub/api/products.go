package api

// GetProductsResponse v1
//
// The response object for getting products.
type GetProductsResponse struct {
	Products []ProductResponse `json:"products"`
}

// ProductResponse v1
//
// The response object for a product.
type ProductResponse struct {
	UID                     string    `json:"uid"`
	Label                   string    `json:"label"`
	AutoProvisionFleets     *[]string `json:"auto_provision_fleets"`
	DisableDevicesByDefault bool      `json:"disable_devices_by_default"`
}

// PostProductRequest v1
//
// The request object for adding a product.
type PostProductRequest struct {
	ProductUID string `json:"product_uid"`
	Label      string `json:"label"`
	// Not required
	AutoProvisionFleets []string `json:"auto_provision_fleets"`
	// Not required
	DisableDevicesByDefault bool `json:"disable_devices_by_default"`
}
