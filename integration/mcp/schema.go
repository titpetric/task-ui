package mcp

// Schema is the JSON schema of a tool's arguments: an object of named
// properties. An empty Properties stays out of the JSON: the schema of an
// array's string elements has none, and a null there makes a client drop
// the tool.
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property is one argument of a tool.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	// Items is the shape of each element of an array.
	Items *Schema `json:"items,omitempty"`
}

// object returns the schema of an object with the properties, the named
// ones required.
func object(properties map[string]Property, required ...string) Schema {
	return Schema{Type: "object", Properties: properties, Required: required}
}
