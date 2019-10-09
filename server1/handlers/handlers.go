package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"crypto/rsa"
	"protected-notebook/server1/file"
	"protected-notebook/server1/idea"
	"protected-notebook/server1/rsa_initial"

	"github.com/gin-gonic/gin"
)

//GetFileListHandler send list of files to client
func GetFileListHandler(c *gin.Context) {
	fmt.Println("Sending list of available files")
	c.JSON(http.StatusOK, file.Get())
}

//GetFileContent send encrypted file content and IDEA key
func GetFileContent(c *gin.Context) {
	fileName := c.Param("name")
	fmt.Println("Finding file...")
	str, err := file.GetContentByName(fileName)
	fmt.Println("Encrypting file content...")
	key, encrypted := idea.CFBEncrypter([]byte(str))
	fmt.Println("Encrypting IDEA key with RSA Public Key...")
	key = rsa_initial.EncryptText(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, gin.H{
		"content": encrypted,
		"key":     key})
}

//AddFileHandler handle adding new files
func AddFileHandler(c *gin.Context) {
	fileItem, statusCode, err := convertHTTPBodyToFile(c.Request.Body)
	if err != nil {
		c.JSON(statusCode, err)
		return
	}
	c.JSON(statusCode, gin.H{
		"id": file.Add(fileItem.Name, fileItem.Path)})
}

//DeleteFileHandler handle deleting file
func DeleteFileHandler(c *gin.Context) {
	fileID := c.Param("id")
	if err := file.Delete(fileID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "")
}

func convertHTTPBodyToFile(httpBody io.ReadCloser) (file.File, int, error) {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return file.File{}, http.StatusInternalServerError, err
	}
	defer httpBody.Close()
	return convertJSONBodyToFile(body)
}

func convertJSONBodyToFile(jsonBody []byte) (file.File, int, error) {
	var fileItem file.File
	err := json.Unmarshal(jsonBody, &fileItem)
	if err != nil {
		return file.File{}, http.StatusBadRequest, err
	}
	return fileItem, http.StatusOK, nil
}

func convertHTTPBodyToString(httpBody io.ReadCloser) (string, int, error) {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	defer httpBody.Close()
	return string(body), http.StatusOK, nil
}

//SetRSAPublicKey setting rsa public key
func SetRSAPublicKey(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	fmt.Println("Setting RSA Public Key...")
	if err != nil {
		panic(err)
	}
	pub := &rsa.PublicKey{}
	err = json.Unmarshal(body, pub)
	if err != nil {
		panic(err)
	}
	rsa_initial.SetRSA(pub)
	GetFileListHandler(c)
}