package brands

// Thing is the base entity, all Public APIs should have these properties
type Thing struct {
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl,omitempty"`
	Types     []string `json:"types,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
}

// Brand represent a brand owned by an organisation, current only used is relation to FT brands
type Brand struct {
	Thing
	Description    string   `json:"description,omitempty"`
	DescriptionXML string   `json:"descriptionXML,omitempty"`
	Strapline      string   `json:"strapline,omitempty"`
	ImageURL       string   `json:"_imageUrl,omitempty"` // NB Temp hack
	Parent         *Thing   `json:"parentBrand,omitempty"`
	Children       []*Thing `json:"childBrands,omitempty"`
}
