package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server is running",
		})
	})

	r.POST("/multiple/", func(c *gin.Context) {
		skuType := c.PostForm("skuType")
		fmt.Println("skuType => ", skuType)

		uploadedUrls := []interface{}{}

		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		files := form.File["files"]
		Url := "http://0.0.0.0:3000/upload-s3/"

		for _, file := range files {
			filename := filepath.Base(file.Filename)
			// fmt.Println("filename =>", filename)

			contentType := file.Header.Values("Content-Type")[0]

			fileData, err := file.Open()
			if err != nil {
				fmt.Println("Unable to get File data")
			}
			var dataBuffer bytes.Buffer
			io.Copy(&dataBuffer, fileData)

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			ioWriter, _ := CreateImageFormFile(writer, filename, contentType)
			ioWriter.Write(dataBuffer.Bytes())
			writer.WriteField("skuType", skuType)
			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, Url, body)

			req.Header.Set("Content-Type", writer.FormDataContentType())

			client := &http.Client{
				Timeout: time.Duration(time.Second * 5),
			}

			resp, err := client.Do(req)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}

			if err != nil {
				fmt.Println("Remote request failed - ", err)
			}

			if resp.StatusCode == http.StatusOK {
				var responseBody map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&responseBody)
				if err != nil {
					fmt.Println(err)
				}
				// fmt.Println(responseBody)
				dataMap := responseBody["data"].(map[string]interface{})
				url := dataMap["url"]
				// fmt.Println(url)
				uploadedUrls = append(uploadedUrls, url)
			} else if len(uploadedUrls) == 0 {
				c.JSON(resp.StatusCode, gin.H{
					"message": "Something Went wrong",
				})
				return
			} else {
				c.JSON(resp.StatusCode, gin.H{
					"message": "Some skus uploaded",
					"Urls":    uploadedUrls,
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"Urls": uploadedUrls,
		})

	})

	r.Run()

}

func CreateImageFormFile(w *multipart.Writer, filename, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

/*
#################### Other Commands #########################

fmt.Printf("file.Header => %#v\n", file.Header.Values("Content-Type")[0])
fmt.Printf("file.Header Type => %T\n", file.Header.Values("Content-Type")[0])


			fmt.Println(os.Getwd())

			out, err := os.Open(file.Filename) // this will work if you place the file in "os.Getwd()" directory
			if err != nil {
				fmt.Println("Unable to open file => ", err)
			}


			fileContents, _ := ioutil.ReadAll(out)

			fmt.Println("fileContents size => ", unsafe.Sizeof(fileContents))

			// fmt.Println("file Size => ", file.Size)

			// body := &bytes.Buffer{}
			// writer := multipart.NewWriter(body)
			// part, _ := writer.CreateFormFile("file", filename)
			// io.Copy(part, out)
			// writer.WriteField("skuType", skuType)

			// writer.Close()
			// fmt.Println(body)
			// ioWriter.Write(fileContents)


*/
