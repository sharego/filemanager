package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	port := 8090
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: filemanager [-h] [-d] [port], default port: %v ©xiaowei\n", port)
	}

	help := flag.Bool("h", false, "help")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	args := flag.Args()
	if len(args) == 1 {
		v, e := strconv.Atoi(args[0])
		if e != nil || v < 1 || v > 65535 {
			flag.Usage()
			return
		}
		port = v
	}

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile(".", true)))

	r.PUT("/*upname", fileUploadHandler)
	r.POST("/*upname", fileUploadHandler)

	fmt.Fprintf(os.Stdout, "Server: 0.0.0.0:%v\n", port)
	err := r.Run("0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server Start failed: %v", err)
		os.Exit(1)
	}
}

func fileUploadHandler(c *gin.Context) {

	dstDir := "."
	filename := ""
	dst := ""
	tmpDst := ""
	tmpSuffix := strconv.FormatInt(time.Now().Unix(), 10) + ".tmp"

	var err error = nil

	if c.Request.Method == "PUT" {
		filename = strings.TrimLeft(c.Request.RequestURI, "/")
		dst = filepath.Join(dstDir, filename)
		tmpDst = dst + tmpSuffix

		out, err := os.Create(tmpDst)
		if err == nil {
			defer out.Close()
			_, err = io.Copy(out, c.Request.Body)
		}

	} else { // POST
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("client upload error: %s\n", err))
			return
		}

		filename = file.Filename

		dst = filepath.Join(dstDir, filename)
		tmpDst = dst + tmpSuffix

		// Upload the file to specific dst.
		err = c.SaveUploadedFile(file, tmpDst)
	}

	defer func() {
		if len(tmpDst) > 0 {
			if _, e := os.Stat(tmpDst); e == nil {
				os.Remove(tmpDst)
			}
		}
	}()

	if err == nil {
		if _, err := os.Stat(dst); err == nil { // file exists
			c.String(http.StatusForbidden, fmt.Sprintf("exists: %s\n", filename))
			return
		}
		err = os.Rename(tmpDst, dst)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("'%s' upload failed: %s\n", filename, err))
	} else {
		c.String(http.StatusOK, fmt.Sprintf("uploaded: %s\n", filename))
	}
}
