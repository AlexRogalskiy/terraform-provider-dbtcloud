package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ConnectionDetails struct {
	Account                string                       `json:"account,omitempty"`
	Database               string                       `json:"database,omitempty"`
	DBName                 string                       `json:"dbname,omitempty"`
	Warehouse              string                       `json:"warehouse,omitempty"`
	AllowSSO               *bool                        `json:"allow_sso,omitempty"`
	ClientSessionKeepAlive *bool                        `json:"client_session_keep_alive,omitempty"`
	Role                   string                       `json:"role,omitempty"`
	OAuthClientID          string                       `json:"oauth_client_id,omitempty"`
	OAuthClientSecret      string                       `json:"oauth_client_secret,omitempty"`
	Host                   string                       `json:"hostname,omitempty"`
	Port                   int                          `json:"port,omitempty"`
	TunnelEnabled          *bool                        `json:"tunnel_enabled,omitempty"`
	AdapterId              *int                         `json:"adapter_id,omitempty"`
	AdapterDetails         *DatabricksCredentialDetails `json:"connection_details,omitempty"`
}

type Connection struct {
	ID                      *int              `json:"id,omitempty"`
	AccountID               int               `json:"account_id"`
	ProjectID               int               `json:"project_id"`
	Name                    string            `json:"name"`
	Type                    string            `json:"type"`
	CreatedByID             *int              `json:"created_by_id,omitempty"`
	CreatedByServiceTokenID *int              `json:"created_by_service_token_id,omitempty"`
	State                   int               `json:"state"`
	PrivateLinkEndpointID   string            `json:"private_link_endpoint_id,omitempty"`
	Created_At              *string           `json:"created_at,omitempty"`
	Updated_At              *string           `json:"updated_at,omitempty"`
	Details                 ConnectionDetails `json:"details"`
}

type ConnectionListResponse struct {
	Data   []Connection   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type ConnectionResponse struct {
	Data   Connection     `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Adapter struct {
	ID                      *int            `json:"id,omitempty"`
	AccountID               int             `json:"account_id"`
	ProjectID               int             `json:"project_id"`
	CreatedByID             *int            `json:"created_by_id,omitempty"`
	CreatedByServiceTokenID *int            `json:"created_by_service_token_id,omitempty"`
	Metadata                AdapterMetadata `json:"metadata_json"`
	State                   int             `json:"state"`
	AdapterVersion          string          `json:"adapter_version"`
	CreatedAt               *string         `json:"created_at,omitempty"`
	UpdatedAt               *string         `json:"updated_at,omitempty"`
}

type AdapterMetadata struct {
	Title     string `json:"title"`
	DocsLink  string `json:"docs_link"`
	ImageLink string `json:"image_link"`
}

type AdapterResponse struct {
	Data   Adapter        `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetConnection(connectionID, projectID string) (*Connection, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) CreateConnection(projectID int, name string, connectionType string, privatelinkEndpointID string, isActive bool, account string, database string, warehouse string, role string, allowSSO *bool, clientSessionKeepAlive *bool, oAuthClientID string, oAuthClientSecret string, hostName string, port int, tunnelEnabled *bool, httpPath string, catalog string) (*Connection, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}

	connectionDetails := ConnectionDetails{}
	if connectionType == "adapter" {
		adapterId, err := c.createDatabricksAdapter(projectID, state)
		if err != nil {
			return nil, err
		}

		connectionDetails.AdapterId = adapterId
		connectionDetails.AdapterDetails = GetDatabricksConnectionDetails(hostName, httpPath, catalog)
	} else {
		connectionDetails.Account = account
		connectionDetails.Warehouse = warehouse
		connectionDetails.Role = role
		connectionDetails.OAuthClientID = oAuthClientID
		connectionDetails.OAuthClientSecret = oAuthClientSecret
		connectionDetails.Host = hostName
		connectionDetails.Port = port
		if connectionType == "snowflake" {
			connectionDetails.Database = database
			connectionDetails.AllowSSO = allowSSO
			connectionDetails.ClientSessionKeepAlive = clientSessionKeepAlive
		} else if connectionType == "redshift" {
			connectionDetails.TunnelEnabled = tunnelEnabled
			connectionDetails.DBName = database
		} else {
			connectionDetails.TunnelEnabled = tunnelEnabled
			connectionDetails.DBName = database
		}
	}
	newConnection := Connection{
		AccountID:             c.AccountID,
		ProjectID:             projectID,
		Name:                  name,
		Type:                  connectionType,
		PrivateLinkEndpointID: privatelinkEndpointID,
		State:                 state,
		Details:               connectionDetails,
	}

	newConnectionData, err := json.Marshal(newConnection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(projectID)), strings.NewReader(string(newConnectionData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	if (oAuthClientID != "") && (oAuthClientSecret != "") {
		connectionResponse.Data.Details.OAuthClientID = oAuthClientID
		connectionResponse.Data.Details.OAuthClientSecret = oAuthClientSecret
	}

	return &connectionResponse.Data, nil
}

func (c *Client) UpdateConnection(connectionID, projectID string, connection Connection) (*Connection, error) {
	connectionData, err := json.Marshal(connection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), strings.NewReader(string(connectionData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) DeleteConnection(connectionID, projectID string) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}

func (c *Client) createDatabricksAdapter(projectID int, state int) (*int, error) {

	newAdapter := Adapter{
		ID:             nil,
		AdapterVersion: "databricks_v0",
		ProjectID:      projectID,
		AccountID:      c.AccountID,
		State:          state,
		Metadata: AdapterMetadata{
			Title:     "Databricks",
			DocsLink:  "https://docs.getdbt.com/reference/warehouse-setups/databricks-setup",
			ImageLink: "https://upload.wikimedia.org/wikipedia/commons/6/63/Databricks_Logo.png",
		},
	}

	currentUser, err := c.GetConnectedUser()
	if err != nil {
		// if GetConnectedUser is the following specific error, it means that the user is using a service token
		// as there is no way to get the current token ID, we always use 1
		if strings.Contains(err.Error(), "This endpoint cannot be accessed with a service token") {
			serviceTokenID := 1
			newAdapter.CreatedByServiceTokenID = &serviceTokenID
		} else {
			// if the error is different, return it
			return nil, err
		}
	} else {
		// if there is no error, the user is using a user token
		newAdapter.CreatedByID = &currentUser.ID
	}

	newAdapterData, err := json.Marshal(newAdapter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/adapters/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(projectID)), strings.NewReader(string(newAdapterData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	adapterResponse := AdapterResponse{}
	err = json.Unmarshal(body, &adapterResponse)
	if err != nil {
		return nil, err
	}

	return adapterResponse.Data.ID, nil
}

func GetDatabricksConnectionDetails(hostName string, httpPath string, catalog string) *DatabricksCredentialDetails {
	noValidation := DatabricksCredentialFieldMetadataValidation{
		Required: false,
	}

	typeMetadata := DatabricksCredentialFieldMetadata{
		Label:        "Connection type",
		Description:  "",
		Field_Type:   "hidden",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	typeField := DatabricksCredentialField{
		Metadata: typeMetadata,
		Value:    "databricks",
	}

	hostMetadata := DatabricksCredentialFieldMetadata{
		Label:        "Server Hostname",
		Description:  "The hostname of the Databricks cluster or SQL warehouse",
		Field_Type:   "text",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	hostField := DatabricksCredentialField{
		Metadata: hostMetadata,
		Value:    hostName,
	}

	httpPathMetadata := DatabricksCredentialFieldMetadata{
		Label:        "HTTP Path",
		Description:  "The HTTP path of the Databricks cluster or SQL warehouse",
		Field_Type:   "text",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	httpPathField := DatabricksCredentialField{
		Metadata: httpPathMetadata,
		Value:    httpPath,
	}

	fieldOrder := []string{"type", "host", "http_path"}
	fields := map[string]DatabricksCredentialField{
		"type":      typeField,
		"host":      hostField,
		"http_path": httpPathField,
	}

	if catalog != "" {
		catalogMetadata := DatabricksCredentialFieldMetadata{
			Label:        "Catalog",
			Description:  "Optional: Catalog name if Unity Catalog is enabled in your Databricks workspace.  Only available in dbt version 1.1 and later",
			Field_Type:   "text",
			Encrypt:      false,
			Overrideable: true,
			Validation:   noValidation,
		}
		catalogField := DatabricksCredentialField{
			Metadata: catalogMetadata,
			Value:    catalog,
		}
		fieldOrder = append(fieldOrder, "catalog")
		fields["catalog"] = catalogField
	}

	databricksCredentialDetails := DatabricksCredentialDetails{
		Fields:      fields,
		Field_Order: fieldOrder,
	}

	return &databricksCredentialDetails
}
