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
	DescriptionXML string  `json:"descriptionXML,omitempty"`
	Strapline      string  `json:"strapline,omitempty"`
	ImageURL       string  `json:"_imageUrl,omitempty"` // NB Temp hack
	Parent         *Thing  `json:"parentBrand,omitempty"`
	Children       []Thing `json:"childBrands,omitempty"`
}

// NeoBrand is the same as Brand, but it receives an extra field and multiple parents.
type NeoBrand struct {
	ID             string     `json:"id,omitempty"`
	Types          []string   `json:"types,omitempty"`
	PrefLabel      string     `json:"prefLabel,omitempty"`
	DescriptionXML string     `json:"descriptionXML,omitempty"`
	Strapline      string     `json:"strapline,omitempty"`
	ImageURL       string     `json:"imageUrl,omitempty"` // NB Temp hack
	Parent         NeoThing   `json:"parent,omitempty"`
	Children       []NeoThing `json:"children,omitempty"`
	Authority      string     `json:"authority,omitempty"`
}

type NeoThing struct {
	ID        string   `json:"id,omitempty"`
	Types     []string `json:"types,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
}

type ConceptApiResponse struct {
	Concept
	ImageURL       string           `json:"imageUrl,omitempty"`
	DescriptionXML string           `json:"descriptionXML,omitempty"`
	Strapline      string           `json:"strapline,omitempty"`
	Broader        []RelatedConcept `json:"broaderConcepts,omitempty"`
	Narrower       []RelatedConcept `json:"narrowerConcepts,omitempty"`
}

type RelatedConcept struct {
	Concept Concept `json:concept,omitempty`
}

type Concept struct {
	ID        string `json:"id,omitempty"`
	ApiURL    string `json:"apiUrl,omitempty"`
	PrefLabel string `json:"prefLabel,omitempty"`
	Type      string `json:"type,omitempty"`
}
