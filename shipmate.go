package shipmate

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

/*
Git clone a repository as bare.  Local cache will not ever be used directly.
Instead, a shallow clone will be installed into each build directory later via
`git clone --depth=1`.
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

	cmd := exec.Command("echo", "git", "fetch", remote)

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

	cmd := exec.Command("echo", "git", "clone", "--depth=1", "file://"+cwd+"/"+repo, root)

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
	cmd := exec.Command("echo", "docker", "build", "-t", repo+":"+rev+suffix, target)

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

func run(remote string, name string, rev string, userDefinedPrefix string, wd string) {
	if err := syncWithRemote(remote, name); err != nil {
		log.Fatal(err)
	}

	/*
	   Walk current working directory and build containers where Dockerfiles exist
	*/
	walkFn := func(pth string, info os.FileInfo, err error) error {
		if strings.HasSuffix(pth, "/Dockerfile") && !strings.Contains(pth, "/"+name+"/") {
			target := path.Dir(pth)

			/*
			   If user specified a prefix, skip paths without it
			*/
			if len(userDefinedPrefix) > 0 && !strings.HasPrefix(strings.TrimPrefix(strings.TrimPrefix(target, wd), "/"), userDefinedPrefix) {
				return nil
			}

			cloneShallowAtLocation(wd, rev, name, target)
			suffix := strings.Replace(strings.TrimPrefix(target, wd), "/", "-", -1)
			executeBuild(rev, name, target, suffix)
		}
		return nil
	}

	filepath.Walk(wd, walkFn)
}
