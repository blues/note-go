package note

// Contact has the basic contact info structure
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type Contact struct {
	Name        string `json:"name,omitempty"`
	Affiliation string `json:"org,omitempty"`
	Role        string `json:"role,omitempty"`
	Email       string `json:"email,omitempty"`
}

// Contacts has contact info for this app
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type Contacts struct {
	Admin *Contact `json:"admin,omitempty"`
	Tech  *Contact `json:"tech,omitempty"`
}
