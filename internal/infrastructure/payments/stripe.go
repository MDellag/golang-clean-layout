package payments

type StripeClient struct {
	apiKey string
}

func NewStripeClient(apiKey string) *StripeClient {
	return &StripeClient{apiKey: apiKey}
}

func (c *StripeClient) CreateCharge(amount int, currency string) (string, error) {
	return "", nil
}
