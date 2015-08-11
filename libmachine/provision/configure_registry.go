package provision

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/registry"
	"github.com/docker/machine/log"
)

type RegistryCommandContext struct {
	ContainerName string
	DockerDir     string
	DockerPort    int
	Ip            string
	Port          string
	AuthOptions   auth.AuthOptions
	RegistryOptions  registry.RegistryOptions
	RegistryImage    string
}

// Wrapper function to generate a docker run registry command (manage or join)
// from a template/context and execute it.
func runRegistryCommandFromTemplate(p Provisioner, cmdTmpl string, registryCmdContext RegistryCommandContext) error {
	var (
		executedCmdTmpl bytes.Buffer
	)

	parsedMasterCmdTemplate, err := template.New("registryCmd").Parse(cmdTmpl)
	if err != nil {
		return err
	}

	parsedMasterCmdTemplate.Execute(&executedCmdTmpl, registryCmdContext)

	log.Debugf("The registry command being run is: %s", executedCmdTmpl.String())

	if _, err := p.SSHCommand(executedCmdTmpl.String()); err != nil {
		return err
	}

	return nil
}

func configureRegistry(p Provisioner, registryOptions registry.RegistryOptions, authOptions auth.AuthOptions) error {
	if !registryOptions.IsRegistry {
		return nil
	}

	ip, err := p.GetDriver().GetIP()
	if err != nil {
		return err
	}

	u, err := url.Parse(registryOptions.Host)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	port := parts[1]

	dockerDir := p.GetDockerOptionsDir()

	registryCmdContext := RegistryCommandContext{
		ContainerName: "",
		DockerDir:     dockerDir,
		DockerPort:    2376,
		Ip:            ip,
		Port:          port,
		AuthOptions:   authOptions,
		RegistryOptions:  registryOptions,
		RegistryImage:    registryOptions.Image,
	}

	// First things first, get the registry image.
	if _, err := p.SSHCommand(fmt.Sprintf("sudo docker pull %s", registryOptions.Image)); err != nil {
		return err
	}

	registryCmdTemplate := `sudo docker run -d \
--restart=always \
--name registry \
-p {{.Port}}:{{.Port}} \
-v {{.DockerDir}}:{{.DockerDir}} \
{{.RegistryImage}} \
manage \
--tlsverify \
--tlscacert={{.AuthOptions.CaCertRemotePath}} \
--tlscert={{.AuthOptions.ServerCertRemotePath}} \
--tlskey={{.AuthOptions.ServerKeyRemotePath}} \
-H {{.RegistryOptions.Host}} \
{{range .RegistryOptions.ArbitraryFlags}} --{{.}}{{end}} {{.RegistryOptions.Discovery}}
`
	log.Debug("Launching docker registry")
	if err := runRegistryCommandFromTemplate(p, registryCmdTemplate, registryCmdContext); err != nil {
		return err
	}	
	return nil
}
