package group

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/messagebird/go-rest-api/contact"

	messagebird "github.com/messagebird/go-rest-api"
)

// Group gets returned by the API.
type Group struct {
	ID       string
	HRef     string
	Name     string
	Contacts struct {
		TotalCount int
		HRef       string
	}
	CreatedDatetime time.Time
	UpdatedDatetime time.Time
}

type GroupList struct {
	Offset     int
	Limit      int
	Count      int
	TotalCount int
	Links      struct {
		First    string
		Previous string
		Next     string
		Last     string
	}
	Items []Group
}

// ListOptions can be used to set pagination options in List() and ListContacts().
type ListOptions struct {
	Limit, Offset int
}

// Request represents a contact for write operations, e.g. for creating a new
// group or updating an existing one.
type Request struct {
	Name string `json:"name"`
}

const (
	// path represents the path to the Groups resource.
	path = "groups"

	// contactPath represents the path to the Contacts resource within Groups.
	contactPath = "contacts"
)

// DefaultListOptions provides reasonable values for List().
var DefaultListOptions = &ListOptions{
	Limit:  10,
	Offset: 0,
}

func Create(c *messagebird.Client, request *Request) (*Group, error) {
	if err := validateCreate(request); err != nil {
		return nil, err
	}

	group := &Group{}

	err := c.Request(group, http.MethodPost, path, request)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func validateCreate(request *Request) error {
	if request.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

// Delete attempts deleting the group with the provided ID. If nil is returned,
// the resource was deleted successfully.
func Delete(c *messagebird.Client, id string) error {
	return c.Request(nil, http.MethodDelete, path+"/"+id, nil)
}

// List retrieves a paginated list of groups, based on the options provided.
// It's worth noting DefaultListOptions.
func List(c *messagebird.Client, options *ListOptions) (*GroupList, error) {
	query, err := listQuery(options)
	if err != nil {
		return nil, err
	}

	groupList := &GroupList{}

	if err := c.Request(groupList, http.MethodGet, path+"?"+query, nil); err != nil {
		return nil, err
	}

	return groupList, nil
}

func listQuery(options *ListOptions) (string, error) {
	if options.Limit < 10 {
		return "", fmt.Errorf("minimum limit is 10, got %d", options.Limit)
	}

	if options.Offset < 0 {
		return "", fmt.Errorf("offset can not be negative")
	}

	values := &url.Values{}

	values.Set("limit", strconv.Itoa(options.Limit))
	values.Set("offset", strconv.Itoa(options.Offset))

	return values.Encode(), nil
}

// Read retrieves the information of an existing group.
func Read(c *messagebird.Client, id string) (*Group, error) {
	group := &Group{}

	err := c.Request(group, http.MethodGet, path+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	return group, nil
}

// Update overrides the group with any values provided in request.
func Update(c *messagebird.Client, id string, groupRequest *Request) error {
	if err := validateUpdate(groupRequest); err != nil {
		return err
	}

	return c.Request(nil, http.MethodPatch, path+"/"+id, groupRequest)
}

func validateUpdate(groupRequest *Request) error {
	if groupRequest.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

// AddContacts adds a maximum of 50 contacts to the group.
func AddContacts(c *messagebird.Client, groupID string, contactIDS []string) error {
	if err := validateAddContacts(contactIDS); err != nil {
		return err
	}

	query := addContactsQuery(contactIDS)

	formattedPath := fmt.Sprintf("%s/%s/%s?%s", path, groupID, contactPath, query)

	return c.Request(nil, http.MethodGet, formattedPath, nil)
}

func validateAddContacts(contactIDS []string) error {
	count := len(contactIDS)

	// len(nil) == 0: https://golang.org/ref/spec#Length_and_capacity
	if count == 0 {
		return fmt.Errorf("contactIDS is required")
	}

	if count > 50 {
		return fmt.Errorf("can not add more than 50 contacts per request, got %d", count)
	}

	return nil
}

// addContactsQuery gets a query string to add contacts to a group. We're using
// the alternative "/foo?_method=PUT&key=value" alternative to send the contact
// IDs as GET params. Sending these in the request body would require a painful
// workaround, as client.Request sends request bodies as JSON by default. See
// also: https://developers.messagebird.com/docs/alternatives.
func addContactsQuery(contactIDS []string) string {
	// Slice's length is one bigger than len(IDs) for the _method param.
	params := make([]string, 0, len(contactIDS)+1)

	params = append(params, "_method="+http.MethodPut)

	for _, contactID := range contactIDS {
		params = append(params, "ids[]="+contactID)
	}

	return strings.Join(params, "&")
}

// ListContacts lists the contacts that are a member of a group.
func ListContacts(c *messagebird.Client, groupID string, options *ListOptions) (*contact.ContactList, error) {
	query, err := listQuery(options)
	if err != nil {
		return nil, err
	}

	formattedPath := fmt.Sprintf("%s/%s/%s?%s", path, groupID, contactPath, query)

	contacts := &contact.ContactList{}

	err = c.Request(contacts, http.MethodGet, formattedPath, nil)

	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// RemoveContact removes the contact from a group. If nil is returned, the
// operation was successful.
func RemoveContact(c *messagebird.Client, groupID, contactID string) error {
	formattedPath := fmt.Sprintf("%s/%s/contacts/%s", path, groupID, contactID)

	return c.Request(nil, http.MethodDelete, formattedPath, nil)
}
