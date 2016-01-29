package brands

// Thing is the base entity, all Public APIs should have these properties
type Thing struct {
	ID        string `json:"id"`
	APIURL    string `json:"apiUrl"`
	PrefLabel string `json:"prefLabel"`
}

// Brand represent a brand owned by an organisation, current only used is relation to FT brands
type Brand struct {
	Thing
	Types          []string `json:"types"`
	Description    string   `json:"description"`
	DescriptionXML string   `json:"descriptionXML"`
	Strapline      string   `json:"strapline"`
	ImageURL       string   `json:"_imageUrl"` // NB Temp hack
	Parent         *Thing   `json:"parentBrand,omitempty"`
}
