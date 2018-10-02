// +build e2e

package threescaleAPI

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/3scale/ostia/threescaleAPI-operator/pkg/threescale/system_client"
)

const (
	testServiceName   = "threescale-e2e-test-service"
	integrationMethod = "apicast"
)

var (
	accessToken    = getAccessToken()
	threescaleHost = getHost()
)

// Removes existing api/service to test in a clean room environment
// Fails if unable to delete service for any reason
func tearDown(t *testing.T, c *client.ThreeScaleClient) {
	api, err := getServiceFromServiceSystemName(c, accessToken, testServiceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		t.Fatalf("error during service teardown")
	}
	err = c.DeleteService(accessToken, api.ID)
	if err != nil {
		t.Fatalf("error deleting existing service. tear down failed. unable to test in clean environment")
	}
}

func createClient() *client.ThreeScaleClient {
	ap, err := client.NewAdminPortal("https", "istiodevel-admin.3scale.net", 443)
	if err != nil {
		panic("error configuring 3scale client admin portal")
	}

	return client.NewThreeScale(ap, http.DefaultClient)
}

func TestEndpointReconciliation(t *testing.T) {
	c := createClient()
	tearDown(t, c)

	// Testing creation
	svc, err := c.CreateService(accessToken, testServiceName)
	if err != nil {
		t.Fatalf("error creating service - %s", err)
	}

	preReconcileEndpoints, err := getEndpointsFrom3scaleSystem(c, accessToken, svc.Name)
	if err != nil {
		t.Fatalf("error parsing desired endpoints from 3scale - %s", err)
	}
	fmt.Println(preReconcileEndpoints)

	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData(generateMockSpec())
	if err != nil {
		t.Fatalf("error loading swagger spec")
	}

	desiredEndpoints, err := getEndpointsFromSwagger(spec)
	fmt.Println(desiredEndpoints)

	if compareEndpoints(preReconcileEndpoints, desiredEndpoints) {
		t.Fatalf("endpoints should not be in desired state")
	}

	err = reconcileEndpointsWith3scaleSystem(c, accessToken, svc, preReconcileEndpoints, desiredEndpoints)
	if err != nil {
		t.Fatalf("error calling reconcile on endpoints")
	}

	reconciledEndpoints, err := getEndpointsFrom3scaleSystem(c, accessToken, svc.Name)
	if err != nil {
		t.Fatalf("error parsing desired endpoints from 3scale post reconciliation - %s", err)
	}

	if !compareEndpoints(reconciledEndpoints, desiredEndpoints) {
		t.Fatalf("endpoints should be in desired state")
	}
}

func getAccessToken() string {
	token := os.Getenv("3SCALE_ACCESS_TOKEN")
	if token == "" {
		panic("End-to-End tests require access token to be present")
	}
	return token
}

func getHost() string {
	token := os.Getenv("3SCALE_HOST")
	if token == "" {
		panic("End-to-End tests require a hos")
	}
	return token
}

func generateMockSpec() []byte {
	spec := []byte(`openapi: 3.0.1
servers:
  - url: '{scheme}://developer.uspto.gov/ds-api'
    variables:
      scheme:
        description: 'The Data Set API is accessible via https and http'
        enum:
          - 'https'
          - 'http'
        default: 'https'
info:
  description: >-
    The Data Set API (DSAPI) allows the public users to discover and search
    USPTO exported data sets. This is a generic API that allows USPTO users to
    make any CSV based data files searchable through API. With the help of GET
    call, it returns the list of data fields that are searchable. With the help
    of POST call, data can be fetched based on the filters on the field names.
    Please note that POST call is used to search the actual data. The reason for
    the POST call is that it allows users to specify any complex search criteria
    without worry about the GET size limitations as well as encoding of the
    input parameters.
  version: 1.0.0
  title: USPTO Data Set API
  contact:
    name: Open Data Portal
    url: 'https://developer.uspto.gov'
    email: developer@uspto.gov
tags:
  - name: metadata
    description: Find out about the data sets
  - name: search
    description: Search a data set
paths:
  /:
    get:
      x-3scale-metrics:
        - metric: hits
          increment: 1
        - metric: root
          increment: 1
      tags:
        - metadata
      operationId: list-data-sets
      summary: List available data sets
      responses:
        '200':
          description: Returns a list of data sets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/dataSetList'
              example:
                {
                  "total": 2,
                  "apis": [
                    {
                      "apiKey": "oa_citations",
                      "apiVersionNumber": "v1",
                      "apiUrl": "https://developer.uspto.gov/ds-api/oa_citations/v1/fields",
                      "apiDocumentationUrl": "https://developer.uspto.gov/ds-api-docs/index.html?url=https://developer.uspto.gov/ds-api/swagger/docs/oa_citations.json"
                    },
                    {
                      "apiKey": "cancer_moonshot",
                      "apiVersionNumber": "v1",
                      "apiUrl": "https://developer.uspto.gov/ds-api/cancer_moonshot/v1/fields",
                      "apiDocumentationUrl": "https://developer.uspto.gov/ds-api-docs/index.html?url=https://developer.uspto.gov/ds-api/swagger/docs/cancer_moonshot.json"
                    }
                  ]
                }
  /{dataset}/{version}/fields:
    get:
      x-3scale-metrics:
        - metric: hits
          increment: 1
        - metric: get_fields
          increment: 1
      tags:
        - metadata
      summary: >-
        Provides the general information about the API and the list of fields
        that can be used to query the dataset.
      description: >-
        This GET API returns the list of all the searchable field names that are
        in the oa_citations. Please see the 'fields' attribute which returns an
        array of field names. Each field or a combination of fields can be
        searched using the syntax options shown below.
      operationId: list-searchable-fields
      parameters:
        - name: dataset
          in: path
          description: 'Name of the dataset.'
          required: true
          example: "oa_citations"
          schema:
            type: string
        - name: version
          in: path
          description: Version of the dataset.
          required: true
          example: "v1"
          schema:
            type: string
      responses:
        '200':
          description: >-
            The dataset API for the given version is found and it is accessible
            to consume.
          content:
            application/json:
              schema:
                type: string
        '404':
          description: >-
            The combination of dataset name and version is not found in the
            system or it is not published yet to be consumed by public.
          content:
            application/json:
              schema:
                type: string
  /{dataset}/{version}/vinils:
    get:
      x-3scale-metrics:
        - metric: hits
          increment: 1
      operationId: get_vinils
  /{dataset}/{version}/records:
    get:
      x-3scale-metrics:
        - metric: hits
          increment: 1
        - metric: get_records
          increment: 1
      operationId: get_records
    post:
      x-3scale-metrics:
        - metric: hits
          increment: 1
        - metric: cacafuti
          increment: 2
        - metric: post_records
          increment: 1
      tags:
        - search
      operationId: perform-search
      parameters:
        - name: version
          in: path
          description: Version of the dataset.
          required: true
          schema:
            type: string
            default: v1
        - name: dataset
          in: path
          description: 'Name of the dataset. In this case, the default value is oa_citations'
          required: true
          schema:
            type: string
            default: oa_citations
      responses:
        '200':
          description: successful operation
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  additionalProperties:
                    type: object
        '404':
          description: No matching record found for the given criteria.
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                criteria:
                  description: >-
                    Uses Lucene Query Syntax in the format of
                    propertyName:value, propertyName:[num1 TO num2] and date
                    range format: propertyName:[yyyyMMdd TO yyyyMMdd]. In the
                    response please see the 'docs' element which has the list of
                    record objects. Each record structure would consist of all
                    the fields and their corresponding values.
                  type: string
                  default: '*:*'
                start:
                  description: Starting record number. Default value is 0.
                  type: integer
                  default: 0
                rows:
                  description: >-
                    Specify number of rows to be returned. If you run the search
                    with default values, in the response you will see 'numFound'
                    attribute which will tell the number of records available in
                    the dataset.
                  type: integer
                  default: 100
              required:
                - criteria
x-3scale-plans:
  plans:
  - name: basic
    limits:
    - max: 100
      metric: hits
      period: month
    - max: 0
      metric: root
      period: eternity
    - max: 10
      metric: hits
      period: minute
    - max: 10
      metric: get_records
      period: month
  - name: premium
    limits:
    - max: 100000
      metric: hits
      period: month
    - max: 1000
      metric: hits
      period: minute
    - max: 100
      metric: get_records
      period: month
components:
  schemas:
    dataSetList:
      type: object
      properties:
        total:
          type: integer
        apis:
          type: array
          items:
            type: object
            properties:
              apiKey:
                type: string
                description: To be used as a dataset parameter value
              apiVersionNumber:
                type: string
                description: To be used as a version parameter value
              apiUrl:
                type: string
                format: uriref
                description: "The URL describing the dataset's fields"
              apiDocumentationUrl:
                type: string
                format: uriref
                description: A URL to the API console for each API
`)
	return spec
}
