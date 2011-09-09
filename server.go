package rest

import (
	"fmt"
	"http"
	"strings"
	"url"
)

var resources = make(map[string]interface{})

// Lists all the items in the resource
// GET /resource/
type index interface {
	Index(http.ResponseWriter)
}

// Lists all the items in the resource
// GET /resource/
type informed_index interface {
	Index(http.ResponseWriter, url.Values, http.Header)
}

// Creates a new resource item
// POST /resource/
type create interface {
	Create(http.ResponseWriter, *http.Request)
}

// Views a resource item
// GET /resource/id
type find interface {
	Find(http.ResponseWriter, string)
}

// Views a resource item
// GET /resource/id
type informed_find interface {
	Find(http.ResponseWriter, string, url.Values, http.Header)
}

// PUT /resource/id
type update interface {
	Update(http.ResponseWriter, string, *http.Request)
}

// Acts on a resource item, or performs a top level action
// POST /resource/id/** or POST /resource/top-level-action
type action interface {
	Act(http.ResponseWriter, []string, *http.Request)
}

type operations interface {
	GetOps(http.ResponseWriter)
	Ops(http.ResponseWriter, string, *http.Request)
}

// DELETE /resource/id
type delete interface {
	Delete(http.ResponseWriter, string)
}

// Return options to use the service. If string is nil, then it is the base URL
// OPTIONS /resource/id
// OPTIONS /resource/
type options interface {
	Options(http.ResponseWriter, string)
}

// Generic resource handler
func resourceHandler(c http.ResponseWriter, req *http.Request) {
	// Parse request URI to resource URI and (potential) ID
	var resourceEnd = strings.Index(req.URL.Path[1:], "/") + 1
	var resourceName string
	if resourceEnd == -1 {
		resourceName = req.URL.Path[1:]
	} else {
		resourceName = req.URL.Path[1:resourceEnd]
	}
	var id = req.URL.Path[resourceEnd+1:]

	resource, ok := resources[resourceName]
	if !ok {
		fmt.Fprintf(c, "resource %s not found\n", resourceName)
	}

	if len(id) == 0 {
		switch req.Method {
		case "GET":
			// Index
			if resIndex, ok := resource.(informed_index); ok {
			    if err := req.ParseForm(); err != nil {
			        BadRequest(c, err.String())
			        return
			    }
                resIndex.Index(c, req.Form, req.Header)
			} else if resIndex, ok := resource.(index); ok {
				resIndex.Index(c)
			} else {
				NotImplemented(c)
			}
		case "POST":
			// Create
			if resCreate, ok := resource.(create); ok {
				resCreate.Create(c, req)
			} else {
				NotImplemented(c)
			}
		case "OPTIONS":
			// automatic options listing
			if resOptions, ok := resource.(options); ok {
				resOptions.Options(c, id)
			} else {
				NotImplemented(c)
			}
		default:
			NotImplemented(c)
		}
	} else { // ID was passed
		switch req.Method {
		case "GET":
			// Find
			if resFind, ok := resource.(informed_find); ok {
			    if err := req.ParseForm(); err != nil {
			        BadRequest(c, err.String())
			        return
			    }
                resFind.Find(c, id, req.Form, req.Header)
            } else if resFind, ok := resource.(find); ok {
				resFind.Find(c, id)
			} else {
				NotImplemented(c)
			}
		case "POST":
			// Action
			if resVerb, ok := resource.(action); ok {
                parts := strings.Split(id, "/")
                if parts[0] == "" {
                    http.Error(c, "invalid uri " + id, http.StatusBadRequest)
                    return
                }

				resVerb.Act(c, parts, req)
			} else {
				NotImplemented(c)
			}
		case "PUT":
			// Update
			if resUpdate, ok := resource.(update); ok {
				resUpdate.Update(c, id, req)
			} else {
				NotImplemented(c)
			}
		case "DELETE":
			// Delete
			if resDelete, ok := resource.(delete); ok {
				resDelete.Delete(c, id)
			} else {
				NotImplemented(c)
			}
		case "OPTIONS":
			// automatic options
			if resOptions, ok := resource.(options); ok {
				resOptions.Options(c, id)
			} else {
				NotImplemented(c)
			}
		default:
			NotImplemented(c)
		}
	}
}

// Add a resource route to http
func Resource(name string, res interface{}) {
	resources[name] = res
	http.Handle("/"+name+"/", http.HandlerFunc(resourceHandler))
}

// Emits a 404 Not Found
func NotFound(c http.ResponseWriter) {
	http.Error(c, "404 Not Found", http.StatusNotFound)
}

// Emits a 501 Not Implemented
func NotImplemented(c http.ResponseWriter) {
	http.Error(c, "501 Not Implemented", http.StatusNotImplemented)
}

// Emits a 201 Created with the URI for the new location
func Created(c http.ResponseWriter, location string) {
	c.Header().Set("Location", location)
	http.Error(c, "201 Created", http.StatusCreated)
}

// Emits a 200 OK with a location. Used when after a PUT
func Updated(c http.ResponseWriter, location string) {
	c.Header().Set("Location", location)
	http.Error(c, "200 OK", http.StatusOK)
}

// Emits a bad request with the specified instructions
func BadRequest(c http.ResponseWriter, instructions string) {
	c.WriteHeader(http.StatusBadRequest)
	c.Write([]byte(instructions))
}

// Emits a 204 No Content
func NoContent(c http.ResponseWriter) {
	http.Error(c, "204 No Content", http.StatusNoContent)
}
