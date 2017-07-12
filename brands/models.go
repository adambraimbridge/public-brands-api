package brands

// Thing is the base entity, all Public APIs should have these properties
type Thing struct {
	ID         string   `json:"id,omitempty"`
	APIURL     string   `json:"apiUrl,omitempty"`
	Types      []string `json:"types,omitempty"`
	DirectType string   `json:"directType,omitempty"`
	PrefLabel  string   `json:"prefLabel,omitempty"`
}

// Brand represent a brand owned by an organisation, current only used is relation to FT brands
type Brand struct {
	Thing
	DescriptionXML string   `json:"descriptionXML,omitempty"`
	Strapline      string   `json:"strapline,omitempty"`
	ImageURL       string   `json:"imageUrl,omitempty"` // NB Temp hack
	Parents        []*Thing `json:"parentBrands,omitempty"`
	Children       []*Thing `json:"childBrands,omitempty"`
}
