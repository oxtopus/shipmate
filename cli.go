package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
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
func cloneBareRepository(remote string, dest string) {
	cmd := exec.Command("git", "clone", "--bare", remote, dest)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	fmt.Println(out.String())
	if err != nil {
		fmt.Println("Failed to clone!")
		log.Fatal(err)
	}
}

/*
Sync local cache with remote, cloning from scratch if necessary.
*/
func syncWithRemote(remote string, dest string) {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s.  Cloning %s.\n", dest, remote)
		cloneBareRepository(remote, dest)
	}

	os.Chdir(dest)
	cmd := exec.Command("git", "fetch", remote)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	fmt.Println(out.String())
	if err != nil {
		log.Fatal(err)
	}
}

/*
Create archive from local bare repository at specific location
*/
func createArchiveAtLocation(rev string, repo string, target string) {
	root := filepath.Join(target, path.Base(repo))

	// Clear out remnants of previous build
	os.RemoveAll(root)
	os.MkdirAll(root, 0700)

	/*
		git archive --remote=REMOTE --format=tar COMMIT | tar -x -C TARGET/NAME
	*/
	archiveCmd := exec.Command("git", "archive", "--remote="+repo, "--format=tar", rev)
	extractCmd := exec.Command("tar", "-x", "-C", root)

	r, w := io.Pipe()
	archiveCmd.Stdout = w
	extractCmd.Stdin = r

	var b2 bytes.Buffer
	extractCmd.Stdout = &b2

	archiveCmd.Start()
	extractCmd.Start()
	archiveCmd.Wait()
	w.Close()
	extractCmd.Wait()
	io.Copy(os.Stdout, &b2)
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
		fmt.Println(in.Text())
	}
	if err := in.Err(); err != nil {
		fmt.Println("error: %s", err)
	}

	cmd.Wait()
	return nil
}

func main() {
	remote := flag.String("remote", "", "Remote repository URL")
	name := flag.String("name", "", "Local destination of bare repository")
	rev := flag.String("rev", "master", "Git revision")

	flag.Parse()

	if len(*name) == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}

	syncWithRemote(*remote, *name)

	walkFn := func(pth string, info os.FileInfo, err error) error {
		if strings.HasSuffix(pth, "/Dockerfile") && !strings.Contains(pth, "/"+*rev+"/") {
			target := path.Dir(pth)
			suffix := strings.Replace(strings.TrimPrefix(target, cwd), "/", "-", -1)
			createArchiveAtLocation(*rev, *name, target)
			executeBuild(*rev, *name, target, suffix)
		}
		return nil
	}

	filepath.Walk(cwd, walkFn)

}