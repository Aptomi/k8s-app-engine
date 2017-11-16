package helm

// Copied from https://github.com/kubernetes/helm/blob/release-2.6/pkg/getter/httpgetter.go

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/helm/pkg/getter"
	"net/http"

	"k8s.io/helm/pkg/tlsutil"
	"k8s.io/helm/pkg/urlutil"
)

// httpGetter is the default HTTP(/S) backend handler
type httpGetter struct {
	client *http.Client
}

// Get performs a Get from repo.Getter and returns the body.
func (g *httpGetter) Get(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	resp, err := g.client.Get(href)
	if err != nil {
		return buf, err
	}
	if resp.StatusCode != 200 {
		return buf, fmt.Errorf("Failed to fetch %s : %s", href, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	resp.Body.Close() // nolint: errcheck
	return buf, err
}

// newHTTPGetter constructs a valid http/https client as Getter
func newHTTPGetter(URL, CertFile, KeyFile, CAFile string) (getter.Getter, error) {
	var client httpGetter
	if CertFile != "" && KeyFile != "" {
		tlsConf, err := tlsutil.NewClientTLS(CertFile, KeyFile, CAFile)
		if err != nil {
			return nil, fmt.Errorf("can't create TLS config for client: %s", err.Error())
		}
		tlsConf.BuildNameToCertificate()

		sni, err := urlutil.ExtractHostname(URL)
		if err != nil {
			return nil, err
		}
		tlsConf.ServerName = sni

		client.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConf,
			},
		}
	} else {
		client.client = http.DefaultClient
	}
	return &client, nil
}

func newGetterProviders() getter.Providers {
	return getter.Providers{
		{
			Schemes: []string{"http", "https"},
			New:     newHTTPGetter,
		},
	}
}
