package letterboxd

type Client interface {
	TestClient(input string) string
}

type letterboxdHttpClient struct {
}

func (l letterboxdHttpClient) TestClient(input string) string {
	return input
}

func NewClient() (Client, error) {
	return letterboxdHttpClient{}, nil
}
