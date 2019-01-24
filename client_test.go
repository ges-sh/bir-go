package bir_test

import (
	"fmt"
	"testing"

	bir "github.com/ges-sh/bir-go"
)

func TestClient(t *testing.T) {
	client := bir.New("abcde12345abcde12345")

	data, err := client.FetchCompanyData("7251801126")
	fmt.Printf("%+v, \n%v", data, err)
}
