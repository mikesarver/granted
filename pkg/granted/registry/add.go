package registry

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	cfflags "github.com/common-fate/granted/pkg/urfav_overrides"

	"github.com/urfave/cli/v2"
)

// Prevent issues where these flags are initialised in some part of the program then used by another part
// For our use case, we need fresh copies of these flags in the app and in the assume command
// we use this to allow flags to be set on either side of the profile arg e.g `assume -c profile-name -r ap-southeast-2`
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "ref", Aliases: []string{"r"}, Usage: "Used to reference a specific commit hash, tag name or branch name"},
	}
}

var AddCommand = cli.Command{
	Name:  "add",
	Flags: GlobalFlags(),
	Action: func(c *cli.Context) error {

		addFlags, err := cfflags.New("assumeFlags", GlobalFlags(), c, 3)
		if err != nil {
			return err
		}

		if c.Args().Len() < 1 {
			return fmt.Errorf("git repository not provided. You need to provide a git repository like 'granted add https://github.com/your-org/your-registry.git'")
		}

		repoURL := c.Args().First()
		fmt.Printf("git clone %s\n", repoURL)

		u, err := url.ParseRequestURI(repoURL)
		if err != nil {
			return errors.New(err.Error())
		}

		repoDirPath, err := GetRegistryLocation(u)
		if err != nil {
			return err
		}

		cmd := exec.Command("git", "clone", repoURL, repoDirPath)

		err = cmd.Run()
		if err != nil {
			// TODO: Will throw an error if the folder already exists and is not an empty directory.
			fmt.Println("the error is", err)
			return err
		}

		fmt.Println("Sucessfully cloned the repo")

		//if a specific ref is passed we will checkout that ref
		fmt.Println("attempting to checkout branch" + addFlags.String("ref"))

		if addFlags.String("ref") != "" {
			fmt.Println("attempting to checkout branch")

			//can be a git hash, tag, or branch name. In that order
			//todo set the path of the repo before checking out
			ref := addFlags.String("ref")
			cmd := exec.Command("git", "checkout", ref)
			cmd.Dir = repoDirPath

			err = cmd.Run()
			if err != nil {
				fmt.Println("the error is", err)
				return err
			}
			fmt.Println("Sucessfully checkout out " + ref)

		}

		if err, ok := isValidRegistry(repoDirPath, repoURL); err != nil || !ok {
			if err != nil {
				return err
			}

			return fmt.Errorf("unable to find `granted.yml` file in %s", repoURL)
		}

		var r Registry
		_, err = r.Parse(repoDirPath)
		if err != nil {
			return err
		}

		// TODO: Run Sync logic here.

		return nil
	},
}

func formatFolderPath(p string) string {
	var formattedURL string = ""

	// remove trailing whitespaces.
	formattedURL = strings.TrimSpace(p)

	// remove trailing '/'
	formattedURL = strings.TrimPrefix(formattedURL, "/")

	// remove .git from the folder name
	formattedURL = strings.Replace(formattedURL, ".git", "", 1)

	return formattedURL
}

func isValidRegistry(folderpath string, url string) (error, bool) {
	dir, err := os.ReadDir(folderpath)
	if err != nil {
		return err, false
	}

	for _, file := range dir {
		if file.Name() == "granted.yml" || file.Name() == "granted.yaml" {
			return nil, true
		}
	}

	return nil, false
}
