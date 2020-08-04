package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"os/exec"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
)

func main() {
	x, _ := zap.NewProductionConfig().Build()
	logger = x.Sugar()
	router := httprouter.New()
	router.GET("/package", getPackage)
	log.Fatal(http.ListenAndServe(":80", router))

}

func getPackage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := uuid.New()
	dir := fmt.Sprintf("/tmp/%s/", id.String())
	must(os.Mkdir(dir, os.ModePerm))
	defer os.RemoveAll(dir)
	v := r.URL.Query()
	pkg := v.Get("package")
	cmd := exec.Command("git", "clone", pkg, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	must(cmd.Run())
	if cmd.ProcessState.ExitCode() != 0 {
		logger.Error("Error running git clone", zap.Int("code", cmd.ProcessState.ExitCode()))
		w.WriteHeader(500)
		return
	}
	cmd = exec.Command("go", "mod", "vendor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	must(cmd.Run())
	if cmd.ProcessState.ExitCode() != 0 {
		logger.Error("Error running go mod vendor", zap.Int("code", cmd.ProcessState.ExitCode()))
		w.WriteHeader(500)
		return
	}
	buf := bytes.NewBuffer([]byte{})
	zipper := zip.NewWriter(buf)
	addFiles(zipper, dir, "")
	err := zipper.Close()
	if err != nil {
		logger.Error("Error running Compression", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	w.Header().Add("content-type", "application/zip")
	w.Header().Add("content-length", fmt.Sprintf("%d", buf.Len()))
	w.WriteHeader(200)
	_, err = io.Copy(w, buf)
	must(err)
	return
}

func must(e error) {
	if e != nil {
		logger.Errorf("bla %v", e)
		panic(e)
	}
}

//lifted from https://stackoverflow.com/questions/37869793/how-do-i-zip-a-directory-containing-sub-directories-or-files-in-golang
func addFiles(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		logger.Infow("Zipping file", "file", fmt.Sprintf("%s%s", basePath, file.Name()))
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			must(err)

			// Add some files to the archive.
			f, err := w.Create(baseInZip + file.Name())
			must(err)
			_, err = f.Write(dat)
			must(err)

		} else if file.IsDir() {

			// Recurse
			newBase := basePath + file.Name() + "/"
			logger.Infow("Recursing and Adding SubDir", "subDir", newBase)

			addFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}
}
