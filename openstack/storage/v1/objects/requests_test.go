package objects

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/testhelper"
)

const (
	tokenId = "abcabcabcabc"
)

var metadata = map[string]string{"Gophercloud-Test": "objects"}

func serviceClient() *gophercloud.ServiceClient {
	return &gophercloud.ServiceClient{
		Provider: &gophercloud.ProviderClient{TokenID: tokenId},
		Endpoint: testhelper.Endpoint(),
	}
}

func TestDownloadObject(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "GET")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successful download with Gophercloud")
	})

	client := serviceClient()
	content, err := Download(client, "testContainer", "testObject", DownloadOpts{}).ExtractContent()
	if err != nil {
		t.Fatalf("Unexpected error downloading object: %v", err)
	}
	if string(content) != "Successful download with Gophercloud" {
		t.Errorf("Expected %s, got %s", "Successful download with Gophercloud", content)
	}
}

func TestListObjectInfo(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "GET")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)

		w.Header().Set("Content-Type", "application/json")
		r.ParseForm()
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			fmt.Fprintf(w, `[{'hash': '451e372e48e0f6b1114fa0724aa79fa1','last_modified': '2009-11-10 23:00:00 +0000 UTC','bytes': 14,'name': 'goodbye','content_type': 'application/octet-stream'}]`)
		default:
			t.Fatalf("Unexpected marker: [%s]", marker)
		}
	})

	client := serviceClient()
	List(client, "testContainer", ListOpts{Full: true}).EachPage(func(page pagination.Page) (bool, error) {
		actual, err := ExtractInfo(page)
		if err != nil {
			t.Errorf("Failed to extract object info: %v", err)
			return false, err
		}

		expected := []Object{
			Object{
				Hash:         "451e372e48e0f6b1114fa0724aa79fa1",
				LastModified: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Bytes:        14,
				Name:         "goodbye",
				ContentType:  "application/octet-stream",
			},
		}

		testhelper.CheckDeepEquals(t, expected, actual)

		return true, nil
	})
}

func TestListObjectNames(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "GET")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "text/plain")

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "")
	})

	client := serviceClient()
	List(client, "testContainer", ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		actual, err := ExtractNames(page)
		if err != nil {
			t.Errorf("Failed to extract object names: %v", err)
			return false, err
		}

		expected := []string{"helloworld", "goodbye"}

		testhelper.CheckDeepEquals(t, expected, actual)

		return true, nil
	})
}

func TestCreateObject(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "PUT")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusCreated)
	})

	client := serviceClient()
	content := bytes.NewBufferString("Did gyre and gimble in the wabe")
	err := Create(client, "testContainer", "testObject", content, CreateOpts{})
	if err != nil {
		t.Fatalf("Unexpected error creating object: %v", err)
	}
}

func TestCopyObject(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "COPY")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		testhelper.TestHeader(t, r, "Destination", "/newTestContainer/newTestObject")
		w.WriteHeader(http.StatusCreated)
	})

	client := serviceClient()
	err := Copy(client, "testContainer", "testObject", CopyOpts{Destination: "/newTestContainer/newTestObject"})
	if err != nil {
		t.Fatalf("Unexpected error copying object: %v", err)
	}
}

func TestDeleteObject(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "DELETE")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	client := serviceClient()
	err := Delete(client, "testContainer", "testObject", DeleteOpts{})
	if err != nil {
		t.Fatalf("Unexpected error deleting object: %v", err)
	}
}

func TestUpateObjectMetadata(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "POST")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		testhelper.TestHeader(t, r, "X-Object-Meta-Gophercloud-Test", "objects")
		w.WriteHeader(http.StatusAccepted)
	})

	client := serviceClient()
	err := Update(client, "testContainer", "testObject", UpdateOpts{Metadata: metadata})
	if err != nil {
		t.Fatalf("Unexpected error updating object metadata: %v", err)
	}
}

func TestGetObject(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	testhelper.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "HEAD")
		testhelper.TestHeader(t, r, "X-Auth-Token", tokenId)
		testhelper.TestHeader(t, r, "Accept", "application/json")
		w.Header().Add("X-Object-Meta-Gophercloud-Test", "objects")
		w.WriteHeader(http.StatusNoContent)
	})

	client := serviceClient()
	expected := metadata
	actual, err := Get(client, "testContainer", "testObject", GetOpts{}).ExtractMetadata()
	if err != nil {
		t.Fatalf("Unexpected error getting object metadata: %v", err)
	}
	testhelper.CheckDeepEquals(t, expected, actual)
}
