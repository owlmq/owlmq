package chord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/owlmq/owlmq/playground/chord/repository"
	"github.com/owlmq/owlmq/playground/utils"
)

func New(ctx *context.Context) Chord_layer {
	return &Chord{
		ctx:        ctx,
		repository: repository.New(),
	}
}

type Chord struct {
	ctx        *context.Context
	repository repository.Repository_layer
}

func (c *Chord) GetFingerTable() *FingerTable {
	return GetFingerTable(c.ctx)
}

func (c *Chord) ShowFingerTable() []string {
	ft, _ := GetFingerTable(c.ctx).GetAll()
	r := []string{}
	for _, v := range ft {
		est := fmt.Sprintf("%s", v.NodeName)
		r = append(r, est)
	}
	return r
}

func (c *Chord) Get(key string) (string, error) {
	if utils.IsKeyInRange(key, (*c.ctx).Value("hostname").(string), (*c.ctx).Value("predecessor").(string)) != true {
		//fmt.Println("GET redirected")
		//empty string because we use the redirectToNextNode as well for the PUT its the value field
		return c.redirectToNextNode(key, "", "GET")
	}
	//fmt.Println("GET not redirected")
	return c.repository.Read(key)
}

func (c *Chord) Put(key string, value string) (err error) {
	if utils.IsKeyInRange(key, (*c.ctx).Value("hostname").(string), (*c.ctx).Value("predecessor").(string)) != true {
		//fmt.Println("PUT redirected")
		_, err := c.redirectToNextNode(key, value, "PUT")
		return err
	}
	//fmt.Println("PUT not redirected")
	return c.repository.Write(key, value)
}

func (c *Chord) redirectToNextNode(key string, value string, method string) (string, error) {
	//get correct value from fingertable
	fe := GetFingerTable(c.ctx).GetOne(key)
	requestURL := filepath.Join(fe.NodeName, "storage", key)
	requestURL = fmt.Sprintf("http://%s", requestURL)
	//fmt.Println("RequestURL", requestURL, " methode", method)

	if method == "GET" {
		res, err := http.Get(requestURL)
		if err != nil {
			log.Printf("error making http request: %s\n", err)
		}
		if res.StatusCode == 404 {
			return "", errors.New("not found")
		}
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("error making http request: %s\n", err)
		}
		return string(resBody), nil
	} else if method == "PUT" {
		req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer([]byte(value)))
		if err != nil {
			log.Printf("error creating http put request: %s\n", err)
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Printf("error executing http put request: %s\n", err)
		}
		if res.StatusCode == 404 {
			return "", errors.New("not found")
		}
		defer res.Body.Close()

		resBody, _ := io.ReadAll(res.Body)
		return string(resBody), nil
	} else {
		log.Print("not defined method in redirectToNextNode")
		return "", errors.New("not a correct method")
	}
}

// joining_node is the node where we want to find the successor and request_node is the node where we ask to find the successor
func (c *Chord) FindSuccessor(joining_node string, request_node string) string {
	if (*c.ctx).Value("successor").(string) == "" {
		return (*c.ctx).Value("hostname").(string)
	}

	type RequestNode struct {
		JoiningNode string `json:"joining_node"`
		Successor   string `json:"successor"`
	}
	var rn RequestNode
	rn.JoiningNode = joining_node

	tosend, err := json.MarshalIndent(rn, "", "    ")
	if err != nil {
		log.Printf(err.Error())
	}
	requestURL := fmt.Sprintf("http://%s/findsuccessor", request_node)
	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBuffer(tosend))
	if err != nil {
		log.Printf("error creating http post request: %s\n", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("error executing http post request: %s\n", err)
	}
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(resBody, &rn); err != nil {
		log.Printf("Failed to Unmarshal: %s\n", err)
	}
	return string(resBody)
}
