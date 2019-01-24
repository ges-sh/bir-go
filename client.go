package bir

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// HTTPClient ...
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is an object which communicates with BIR
type Client struct {
	apiKey  string
	client  HTTPClient
	baseURL *url.URL

	// sid is session id, which is used for request authentication.
	sid string
	// lastLogin specifies how old sid is, so we know when to refresh it
	lastLogin time.Time
}

var birTestEndpoint = mustParseURL("https://Wyszukiwarkaregontest.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc")

var birEndpoint = mustParseURL("https://Wyszukiwarkaregon.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc")

// New returns new Client
func New(apiKey string) Client {
	return NewWithClient(apiKey, &http.Client{})
}

// NewWithClient returns new Client with custom HTTP client
func NewWithClient(apiKey string, client HTTPClient) Client {
	c := Client{
		apiKey:  apiKey,
		client:  client,
		baseURL: birTestEndpoint,
	}

	if apiKey != "abcde12345abcde12345" {
		c.baseURL = birEndpoint
	}

	return c
}

func mustParseURL(addr string) *url.URL {
	url, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	return url
}

// CompanyData contains data about company
type CompanyData struct {
	Data Data `xml:"dane"`
}

// Data contains company data fields
type Data struct {
	Regon     string `xml:"Regon"`
	Name      string `xml:"Nazwa"`
	State     string `xml:"Wojewodztwo"`
	County    string `xml:"Powiat"`
	Community string `xml:"Gmina"`
	City      string `xml:"Miejscowosc"`
	PostCode  string `xml:"KodPocztowy"`
	Street    string `xml:"Ulica"`
}

// FetchCompanyData fetches company data from BIR register
func (c Client) FetchCompanyData(nip string) (CompanyData, error) {
	// if sid is older than 55 minutes, refresh
	if time.Since(c.lastLogin) > time.Minute*55 {
		err := c.refreshSid()
		if err != nil {
			return CompanyData{}, err
		}
	}

	body := &bytes.Buffer{}

	template.Must(template.New("search").Parse(searchEnvelope)).Execute(body, search{
		Nip: nip,
	})

	request, err := http.NewRequest(http.MethodGet, c.baseURL.String(), body)
	if err != nil {
		return CompanyData{}, err
	}

	request.Header.Set("Content-Type", "application/soap+xml;charset=utf-8")
	request.Header.Set("SOAPAction", "DaneSzukaj")
	request.Header.Set("sid", c.sid)

	resp, err := c.client.Do(request)
	if err != nil {
		return CompanyData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return CompanyData{}, fmt.Errorf("invalid response status code: %v", resp.StatusCode)
	}

	return findCompanyData(resp.Body)
}

var companyDataRe = regexp.MustCompile(`(?s)<DaneSzukajResult>(.+)</DaneSzukajResult>`)

func findCompanyData(r io.Reader) (CompanyData, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return CompanyData{}, err
	}

	data := companyDataRe.FindSubmatch(b)
	if data == nil {
		return CompanyData{}, errors.New("missing company data")
	}

	rawData := html.UnescapeString(string(data[1]))

	var companyData CompanyData
	err = xml.NewDecoder(strings.NewReader(rawData)).Decode(&companyData)
	return companyData, err
}

// login should only be used when client sid expired
func (c *Client) refreshSid() error {
	body := &bytes.Buffer{}

	// TODO keep parsed templates in client or init?
	template.Must(template.New("name").Parse(loginEnvelope)).Execute(body, login{
		APIKey: c.apiKey,
	})

	request, err := http.NewRequest(http.MethodGet, c.baseURL.String(), body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/soap+xml;charset=utf-8")
	request.Header.Set("SOAPAction", "Zaloguj")

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response status code: %v", resp.StatusCode)
	}

	c.sid, err = findSid(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

var loginResult = regexp.MustCompile("<ZalogujResult>")

// findSid is looking for sid in the login response body
func findSid(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	loc := loginResult.FindIndex(b)
	if loc == nil {
		return "", errors.New("sid not found")
	}

	// sid has the length of 20 characters (alphanumeric)
	sid := b[loc[1] : loc[1]+20]
	return string(sid), nil
}
