package listing

import (
	"github.com/davidakpele/property-marketplace/internal/listing/application"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
)

type Service struct {
	Create *application.CreateListingUseCase
	Get    *application.GetListingUseCase
	List   *application.SearchListingsUseCase
	Update *application.UpdateListingUseCase
	Delete *application.DeleteListingUseCase
	Search *application.SearchListingsUseCase
}

func NewService(repo repository.ListingRepository) *Service {
	return &Service{
		Create: application.NewCreateListingUseCase(repo),
		Get:    application.NewGetListingUseCase(repo),
		List:   application.NewSearchListingsUseCase(repo),
		Update: application.NewUpdateListingUseCase(repo),
		Delete: application.NewDeleteListingUseCase(repo),
		Search: application.NewSearchListingsUseCase(repo),
	}
}
