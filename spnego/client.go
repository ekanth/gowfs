package spnego

import (
	"net/url"
	"net/http"
	"os/user"
	"io"
	"fmt"
    "gopkg.in/jcmturner/gokrb5.v4/client"
	"gopkg.in/jcmturner/gokrb5.v4/config"
	"gopkg.in/jcmturner/gokrb5.v4/credentials"
	"gopkg.in/jcmturner/gokrb5.v4/keytab"
	"strings"
)

type Client struct {
	client        http.Client
	userName      string
	keytabPath    string
}

func NewSpnegoClient(userName string, keytabPath string, transport *http.Transport) *Client {
	httpClient := http.Client {
		Transport: transport,
	}

	spnegoClient := Client {
		client: httpClient,
		userName: userName,
		keytabPath: keytabPath,
	}

	return &spnegoClient
}

func (c *Client) setSPNEGOHeader(req *http.Request) error {
	var cl client.Client
	var u *user.User
	var err error

	if c.userName == "" {
		u, err = user.Current()
		if err != nil {
			return fmt.Errorf("Error getting current user: %v", err)
		}
	} else {
		u, err = user.Lookup(c.userName)
		if err != nil {
			return fmt.Errorf("Error looking up user (%s): %v", c.userName, err)
		}
	}
	cfg, _ := config.Load("/etc/krb5.conf")

	if c.keytabPath != "" {
		kt, err := keytab.Load(c.keytabPath)
		if err != nil {
			return fmt.Errorf("Errory loading keytab: %v", err)
		}
		cl = client.NewClientWithKeytab(u.Username, "", kt)
	} else {
		var krbCacheFile string = "/tmp/krb5cc_" + u.Uid
		ccache, err := credentials.LoadCCache(krbCacheFile)
		if err != nil {
			return fmt.Errorf("Error loading kerberos tgt from cache file: %v", err)
		}
		fmt.Println("Principal Name: ", ccache.DefaultPrincipal.PrincipalName.NameString)
		cl, err = client.NewClientFromCCache(ccache)
		if err != nil {
			return fmt.Errorf("Error creating client from kerberos tgt in cache: ", err)
		}
	}
	cl.WithConfig(cfg)

	cl.SetSPNEGOHeader(req, "")

	return nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.setSPNEGOHeader(req)
	return c.client.Do(req)
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
