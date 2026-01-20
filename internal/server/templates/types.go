package templates

// GenericTemplate represents the input data for creating a new template
type GenericTemplate struct {
	Id              string      `json:"id" jsonschema_description:"The id of the template, should be lowercased and separated by underscores."`
	Name            string      `json:"name" jsonschema_description:"The name of the template, the same as the id, but with each word capitalized and replace the underscores with spaces."`
	Parameters      []Parameter `json:"parameters" jsonschema_description:"The parameters that will be placed inside the template.yaml to customize each deployment."`
	Description     string      `json:"description" jsonschema_description:"The description of the template."`
	UpdatePackages  bool        `json:"update_packages" jsonschema_description:"If true will update all the packages."`
	UpgradePackages bool        `json:"upgrade_packages" jsonschema_description:"If true will upgrade all the packages."`
	Packages        []string    `json:"packages" jsonschema_description:"The packages to install on the system."`
	Commands        []string    `json:"commands" jsonschema_description:"The commands to run when the system is installed."`
	Files           []File      `json:"files" jsonschema_description:"Specify the files that needs to be available on the system, such as config files and other files needed by the installed packages and applications."`
}

// Parameter represents a template parameter definition
type Parameter struct {
	Name        string `json:"name" jsonschema_description:"The name of the parameter, needs to be written in Pascal case. If include it in template.yaml as templates needs to be done conform to Go html/template conventions."`
	Description string `json:"description" jsonschema_description:"The description about what the parameter is about."`
}

// File represents a file to be written on the system
type File struct {
	Path    string `json:"path" jsonschema_description:"The path where the file will be created on the system."`
	Content string `json:"content" jsonschema_description:"The content of the files that will be written to the system."`
}

// Template represents the content files of a template
type Template struct {
	Description Description
	Content     string
}

// Description represents the metadata of a template
type Description struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
}
