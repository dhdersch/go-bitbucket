package bitbucket

import (
	"github.com/dhdersch/mapstructure"
)

type Workspace struct {
	c *Client

	Repositories *Repositories
	Permissions  *Permission

	UUID       string
	Type       string
	Slug       string
	Is_Private bool
	Name       string
}

type WorkspaceList struct {
	Page       int
	Pagelen    int
	MaxDepth   int
	Size       int
	Next       string
	Workspaces []Workspace
}

type Permission struct {
	c *Client

	Type string
}

func (t *Permission) GetUserPermissions(organization, member string) (*Permission, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/permissions?q=user.nickname=\"%s\"", organization, member)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodePermission(response), err
}

func (t *Permission) GetUserPermissionsByUuid(organization, member string) (*Permission, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/permissions?q=user.uuid=\"%s\"", organization, member)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodePermission(response), err
}

func (t *Workspace) List() (*WorkspaceList, error) {
	urlStr := t.c.requestUrl("/workspaces")
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeWorkspaceList(response)
}

func (t *Workspace) Get(workspace string) (*Workspace, error) {
	urlStr := t.c.requestUrl("/workspaces/%s", workspace)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeWorkspace(response)
}

func (w *Workspace) Members(teamname string) (interface{}, error) {
	urlStr := w.c.requestUrl("/workspaces/%s/members", teamname)
	return w.c.execute("GET", urlStr, "")
}

func (w *Workspace) Projects(teamname string) (interface{}, error) {
	urlStr := w.c.requestUrl("/workspaces/%s/projects/", teamname)
	return w.c.execute("GET", urlStr, "")
}

func decodePermission(permission interface{}) *Permission {
	permissionResponseMap := permission.(map[string]interface{})
	if permissionResponseMap["size"].(float64) == 0 {
		return nil
	}

	permissionValues := permissionResponseMap["values"].([]interface{})
	if len(permissionValues) == 0 {
		return nil
	}

	permissionValue := permissionValues[0].(map[string]interface{})
	return &Permission{
		Type: permissionValue["permission"].(string),
	}
}

func decodeWorkspace(workspace interface{}) (*Workspace, error) {
	var workspaceEntry Workspace
	workspaceResponseMap := workspace.(map[string]interface{})

	if workspaceResponseMap["type"] != nil && workspaceResponseMap["type"] == "error" {
		return nil, DecodeError(workspaceResponseMap)
	}

	err := mapstructure.Decode(workspace, &workspaceEntry)
	return &workspaceEntry, err
}

func decodeWorkspaceList(workspaceResponse interface{}) (*WorkspaceList, error) {
	workspaceResponseMap := workspaceResponse.(map[string]interface{})
	workspaceMapList := workspaceResponseMap["values"].([]interface{})

	var workspaces []Workspace
	for _, workspaceMap := range workspaceMapList {
		workspaceEntry, err := decodeWorkspace(workspaceMap)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, *workspaceEntry)
	}

	page, ok := workspaceResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := workspaceResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := workspaceResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := workspaceResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := workspaceResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	workspacesList := WorkspaceList{
		Page:       int(page),
		Pagelen:    int(pagelen),
		MaxDepth:   int(max_depth),
		Size:       int(size),
		Next:       next,
		Workspaces: workspaces,
	}

	return &workspacesList, nil
}
