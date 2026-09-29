package handler

type CreateAgentRequest struct {
	Name   string `json:"name"   validate:"required,min=2,max=255" example:"Ada Okafor"`
	Email  string `json:"email"  validate:"required,email"          example:"ada@realty.ng"`
	Phone  string `json:"phone"  validate:"required,min=7,max=50"   example:"+2348011111111"`
	Agency string `json:"agency" validate:"required,min=2,max=255"  example:"Realty NG"`
}

type UpdateAgentRequest struct {
	Name   string `json:"name"   validate:"required,min=2,max=255" example:"Ada Okafor"`
	Email  string `json:"email"  validate:"required,email"          example:"ada@realty.ng"`
	Phone  string `json:"phone"  validate:"required,min=7,max=50"   example:"+2348011111111"`
	Agency string `json:"agency" validate:"required,min=2,max=255"  example:"Realty NG"`
}
