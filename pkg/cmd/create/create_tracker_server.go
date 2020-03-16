package create

import (
	"fmt"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/cmd/create/options"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/jenkins-x/jx/pkg/cmd/helper"

	"github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/cmd/templates"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
)

var (
	createTrackerServer_long = templates.LongDesc(`
		Adds a new Issue Tracker Server URL
`)

	createTrackerServer_example = templates.Examples(`
		# Add a new issue tracker server URL
		jx create tracker server jira myURL
	`)

	trackerKindToServiceName = map[string]string{
		"bitbucket": "bitbucket-bitbucket",
	}
)

// CreateTrackerServerOptions the options for the create spring command
type CreateTrackerServerOptions struct {
	options.CreateOptions

	Name string
}

// NewCmdCreateTrackerServer creates a command object for the "create" command
func NewCmdCreateTrackerServer(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &CreateTrackerServerOptions{
		CreateOptions: options.CreateOptions{
			CommonOptions: commonOpts,
		},
	}

	cmd := &cobra.Command{
		Use:     "server kind [url] [username]",
		Short:   "Creates a new issue tracker server URL",
		Aliases: []string{"provider"},
		Long:    createTrackerServer_long,
		Example: createTrackerServer_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			helper.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.Name, "name", "n", "", "The name for the issue tracker server being created")
	return cmd
}

// Run implements the command
func (o *CreateTrackerServerOptions) Run() error {
	args := o.Args
	trackerUser := ""
	if len(args) < 1 {
		return missingTrackerArguments()
	}
	kind := args[0]
	name := o.Name
	if name == "" {
		name = kind
	}
	gitUrl := ""
	if len(args) > 1 {
		gitUrl = args[1]
	} else {
		// lets try find the git URL based on the provider
		serviceName := trackerKindToServiceName[kind]
		if serviceName != "" {
			url, err := o.FindService(serviceName)
			if err != nil {
				return fmt.Errorf("Failed to find %s issue tracker serivce %s: %s", kind, serviceName, err)
			}
			gitUrl = url
		}
	}

	if gitUrl == "" {
		return missingTrackerArguments()
	}
	authConfigSvc, err := o.CreateIssueTrackerAuthConfigService(kind)
	if err != nil {
		return err
	}
	if len(args) > 2 && kind == "jira" {
		trackerUser = args[2]
		o.Username = trackerUser
		trackerToken := ""
		trackerToken, apiToken, bearerToken, password := "", "", "", ""

		prompt := &survey.Input{
			Message: "issue tracker API Token",
			Default: "",
			Help:    "API Authentication token for the issue tracker",
		}
		showPromptIfOptionNotSet(&trackerToken, prompt, o.In, o.Out, o.Err)
		apiToken = trackerToken
		password = trackerToken

		o.OAUTHToken = trackerToken

		authConfigSvc.SaveUserAuth(gitUrl, &auth.UserAuth{trackerUser, apiToken,bearerToken,password, ""} )
		log.Logger().Infof("Added user %s for server %s with URL %s", trackerUser, util.ColorInfo(name), util.ColorInfo(gitUrl))
	}
	config := authConfigSvc.Config()
	config.GetOrCreateServerName(gitUrl, name, kind)
	config.CurrentServer = gitUrl
	err = authConfigSvc.SaveConfig()
	if err != nil {
		return err
	}
	log.Logger().Infof("Added issue tracker server %s for URL %s", util.ColorInfo(name), util.ColorInfo(gitUrl))
	return nil
}

func missingTrackerArguments() error {
	return fmt.Errorf("Missing tracker server URL arguments. Usage: jx create tracker server kind [url]")
}
