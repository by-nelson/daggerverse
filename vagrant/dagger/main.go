// A simple module to generate a Vagrant Cloud upload url

// This module can be use to create a Vagrant Cloud box, version and provider,
// and it can return the url to upload your local vagrant box. This module is a chain of REST api calls
// to the Vagrant Cloud hence there are no special requirements to use it other than a Vagrant Cloud token.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	boxesApi     = "https://app.vagrantup.com/api/v2/boxes"
	versionsApi  = "https://app.vagrantup.com/api/v2/box/%s/%s/versions"
	providersApi = "https://app.vagrantup.com/api/v2/box/%s/%s/version/%s/providers"

	getApi         = "https://app.vagrantup.com/api/v2/box/%s/%s"
	getVersionApi  = "https://app.vagrantup.com/api/v2/box/%s/%s/version/%s"
	getProviderApi = "https://app.vagrantup.com/api/v2/box/%s/%s/version/%s/provider/%s/%s"
	getUploadApi   = "https://app.vagrantup.com/api/v2/box/%s/%s/version/%s/provider/%s/%s/upload"
)

type Vagrant struct {
	Token    string
	Response string

	Box *Box
}

type Box struct {
	Username     string
	Name         string
	Short        string
	Description  string
	Provider     string
	Architecture string
	Version      string

	Private bool
}

type Upload struct {
	Path string `json:"upload_path"`
}

func newBox(username, boxname string) *Box {
	return &Box{
		Username: username,
		Name:     boxname,
		Private:  false,
	}
}

// Box REST calls body
func (b *Box) createData() string {
	if b.Private {
		return fmt.Sprintf(`{ "box": { "username": "%s", "name": "%s", "is_private": true } }`, b.Username, b.Name)
	} else {
		return fmt.Sprintf(`{ "box": { "username": "%s", "name": "%s", "is_private": false } }`, b.Username, b.Name)
	}
}

func (b *Box) providerData() string {
	return fmt.Sprintf(`{ "provider": { "name": "%s", "architecture": "%s"} }`, b.Provider, b.Architecture)
}

func (b *Box) versionData() string {
	return fmt.Sprintf(`{ "version": { "version": "%s"} }`, b.Version)
}

// Box REST calls endpoints
func (b *Box) createEndpoint() string {
	return boxesApi
}

func (b *Box) versionsEndpoint() string {
	return fmt.Sprintf(versionsApi, b.Username, b.Name)
}

func (b *Box) providersEndpoint() string {
	return fmt.Sprintf(providersApi, b.Username, b.Name, b.Version)
}

func (b *Box) getEndpoint() string {
	return fmt.Sprintf(getApi, b.Username, b.Name)
}

func (b *Box) getVersionEndpoint() string {
	return fmt.Sprintf(getVersionApi, b.Username, b.Name, b.Version)
}

func (b *Box) getProviderEndpoint() string {
	return fmt.Sprintf(getProviderApi, b.Username, b.Name, b.Version, b.Provider, b.Architecture)
}

func (b *Box) getUploadEndpoint() string {
	return fmt.Sprintf(getUploadApi, b.Username, b.Name, b.Version, b.Provider, b.Architecture)
}

// private function
func (m *Vagrant) setHeaders() http.Header {

	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	return headers
}

// private gets
func (m *Vagrant) getBox(ctx context.Context) (*Vagrant, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.Box.getEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.getEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

func (m *Vagrant) getVersion(ctx context.Context) (*Vagrant, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.Box.getVersionEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.getVersionEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

func (m *Vagrant) getProvider(ctx context.Context) (*Vagrant, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.Box.getProviderEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.getProviderEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

// gets upload path
func (m *Vagrant) getUpload(ctx context.Context) (*Upload, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.Box.getUploadEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.getUploadEndpoint(), response.Status)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create response body reader")
	}

	var upload Upload
	err = json.Unmarshal(body, &upload)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get upload data.")
	}

	return &upload, nil
}

// Creates a new vagrant box
func (m *Vagrant) CreateBox(
	ctx context.Context,
	// Vagrant Cloud Username
	usernameArg string,
	// Box name
	boxnameArg string,
	// Vagrant Cloud User's token
	tokenArg string,
) (*Vagrant, error) {

	m.Token = tokenArg
	m.Box = newBox(usernameArg, boxnameArg)

	// Check if box exists
	exists, _ := m.getBox(ctx)

	// Box already created, continue
	if exists != nil {
		return m, nil
	}

	// Box does not exist, creating
	reader := strings.NewReader(m.Box.createData())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Box.createEndpoint(), reader)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.createEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

// Creates a new box version
func (m *Vagrant) WithVersion(
	ctx context.Context,
	// Box version to create
	versionArg string,
) (*Vagrant, error) {

	if m.Box == nil {
		return nil, fmt.Errorf("Needs to be called after create-box.")
	}

	m.Box.Version = versionArg
	reader := strings.NewReader(m.Box.versionData())

	// Check if box exists
	exists, err := m.getBox(ctx)
	if err != nil {
		return nil, err
	}

	// Box not created, exit
	if exists == nil {
		return nil, fmt.Errorf("No exiting box found, create one with create-box")
	}

	// Check if version exists
	exists, _ = m.getVersion(ctx)

	// Version created, continue
	if exists != nil {
		return exists, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Box.versionsEndpoint(), reader)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.versionsEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

// Creates a new box version provider
func (m *Vagrant) WithProvider(
	ctx context.Context,
	// Box provider to create. E.g. virtualbox
	providerArg,
	// Box architecture to use. E.g. amd64
	architectureArg string,
) (*Vagrant, error) {

	if m.Box == nil {
		return nil, fmt.Errorf("Needs to be called after create-box.")
	}

	m.Box.Provider = providerArg
	m.Box.Architecture = architectureArg
	reader := strings.NewReader(m.Box.providerData())

	// Check if box exists
	exists, err := m.getBox(ctx)
	if err != nil {
		return nil, err
	}

	// Box not created, exit
	if exists == nil {
		return nil, fmt.Errorf("No exiting box found, create one with create-box")
	}

	// Check if version exists
	exists, err = m.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	// Version not created, exit
	if exists == nil {
		return nil, fmt.Errorf("No exiting version found, create one with create-box with-version")
	}

	// Check if provider exists
	exists, _ = m.getProvider(ctx)

	// Provider created, continue
	if exists != nil {
		return exists, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Box.providersEndpoint(), reader)
	if err != nil {
		return nil, err
	}

	request.Header = m.setHeaders()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got a failing response for %s: %s.", m.Box.providersEndpoint(), response.Status)
	}

	m.Response = fmt.Sprintf("%v", response)
	return m, nil
}

// Generates an upload url from a provider
func (m *Vagrant) Upload(ctx context.Context) (string, error) {
	if m.Box == nil {
		return "", fmt.Errorf("Needs to be called after create-box.")
	}

	// Check if box exists
	exists, err := m.getBox(ctx)
	if err != nil {
		return "", err
	}

	// Box not created, exit
	if exists == nil {
		return "", fmt.Errorf("No exiting box found, create one with create-box")
	}

	// Get upload path
	upload, err := m.getUpload(ctx)
	if err != nil {
		return "", err
	}

	return upload.Path, nil
}

// Prints last http response
func (m *Vagrant) Debug(ctx context.Context) string {
	return m.Response
}
