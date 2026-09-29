package domain

type ListingType string

const (
	ListingTypeRent     ListingType = "rent"
	ListingTypeSale     ListingType = "sale"
	ListingTypeShortlet ListingType = "shortlet"
)

func (t ListingType) IsValid() bool {
	switch t {
	case ListingTypeRent, ListingTypeSale, ListingTypeShortlet:
		return true
	}
	return false
}
