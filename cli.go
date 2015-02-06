package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

/*
Git clone a repository as bare.  Local cache will not ever be used directly.
Instead, an archive will be installed into each build directory later via
`git archive`
*/
func cloneBareRepository(remote string, dest string) error {
	cmd := exec.Command("git", "clone", "--bare", remote, dest)
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

/*
Sync local cache with remote, cloning from scratch if necessary.
*/
func syncWithRemote(remote string, dest string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		cloneBareRepository(remote, dest)
	}

	if err := os.Chdir(dest); err != nil {
		return err
	}

	cmd := exec.Command("git", "fetch", remote)

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

/*
Create shallow clone at specified location
*/
func cloneShallowAtLocation(cwd string, rev string, repo string, target string) error {
	root := filepath.Join(target, path.Base(repo))

	// Clear out remnants of previous build
	os.RemoveAll(root)
	os.MkdirAll(root, 0700)

	cmd := exec.Command("git", "clone", "--depth=1", "file://"+cwd+"/"+repo, root)

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

/*
Execute docker build
*/
func executeBuild(rev string, repo string, target string, suffix string) error {
	cmd := exec.Command("docker", "build", "-t", repo+":"+rev+suffix, target)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	in := bufio.NewScanner(stdout)

	for in.Scan() {
		log.Println(in.Text())
	}

	if err := in.Err(); err != nil {
		return err
	}

	cmd.Wait()
	return nil
}

func main() {
	remote := flag.String("remote", "", "Remote repository URL")
	name := flag.String("name", "", "Local destination of bare repository")
	rev := flag.String("rev", "master", "Git revision")
	userDefinedPrefix := flag.String("prefix", "", "Limit builds to paths with this prefix.  If not specified, process all.")

	flag.Parse()

	if len(*name) == 0 || len(*remote) == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if err := syncWithRemote(*remote, *name); err != nil {
		log.Fatal(err)
	}

	/*
		Walk current working directory and build containers where Dockerfiles exist
	*/
	walkFn := func(pth string, info os.FileInfo, err error) error {
		if strings.HasSuffix(pth, "/Dockerfile") && !strings.Contains(pth, "/"+*name+"/") {
			target := path.Dir(pth)

			if len(*userDefinedPrefix) > 0 && !strings.HasPrefix(strings.TrimPrefix(strings.TrimPrefix(target, cwd), "/"), *userDefinedPrefix) {
				return nil
			}

			suffix := strings.Replace(strings.TrimPrefix(target, cwd), "/", "-", -1)
			cloneShallowAtLocation(cwd, *rev, *name, target)
			executeBuild(*rev, *name, target, suffix)
		}
		return nil
	}

	filepath.Walk(cwd, walkFn)
}
