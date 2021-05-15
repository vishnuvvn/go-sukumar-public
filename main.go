package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/gofiber/fiber/v2"
)

var (
	atrk  string = "atrk"
	atrv  string = "atrv"
	atrt  string = "atrt"
	uatrk string = "uatrk"
	uatrv string = "uatrv"
	uatrt string = "uatrt"
)

const (
	cType              = "Content-Type"
	application        = "application/json"
	webhookURL         = "https://webhook.site/9628abaf-17c1-4865-a93e-c657ef8b70b2"
	successMsg  string = "{\"status\" : \"Success\"}"
	failMsg     string = "{\"status\" : \"Failed\",\"Reason\" :\"invalid JSON received\"}"
	attributes  string = "attributes"
	value       string = "value"
	ttype       string = "type"
	traits      string = "traits"
)

var mg = sync.WaitGroup{}

func main() {

	app := fiber.New()
	ch := make(chan []byte)
	lookupTable := map[string]string{
		"ev":  "event",
		"et":  "event_type",
		"id":  "app_id",
		"uid": "user_id",
		"mid": "message_id",
		"t":   "page_title",
		"p":   "page_url",
		"l":   "browser_language",
		"cs":  "screen_size",
	}

	app.Post("/postjson", func(c *fiber.Ctx) error {
		var str string
		mg.Add(1)
		flag := isJSON(c.Body())
		if flag {
			go processJson(ch, lookupTable)
			ch <- c.Body()
			str = successMsg
		} else {
			str = failMsg
		}
		c.Response().Header.Set(cType, application)
		return c.SendString(str)
	})
	mg.Wait()
	app.Listen(":3000")
}

func processJson(ch <-chan []byte, lookuptable map[string]string) {

	body := generateResJSON(<-ch, lookuptable)
	uploadJSON(body)
	mg.Done()
}

func uploadJSON(body []byte) {
	url := webhookURL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set(cType, application)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func generateResJSON(data []byte, lookuptable map[string]string) []byte {

	jsonObj := gabs.New()
	jsonParsed, err := gabs.ParseJSON(data)
	if err != nil {
		panic(err)
	}

	//attribute conversion
	for i := 1; true; i++ {
		att_key := fmt.Sprintf("%s%d", atrk, i)
		att_value := fmt.Sprintf("%s%d", atrv, i)
		att_type := fmt.Sprintf("%s%d", atrt, i)

		if jsonParsed.Exists(att_key) && jsonParsed.Exists(att_value) && jsonParsed.Exists(att_type) {
			pathkey := jsonParsed.Search(att_key).Data().(string)
			jsonObj.Set(jsonParsed.Search(att_type), attributes, pathkey, ttype)
			jsonObj.Set(jsonParsed.Search(att_value), attributes, pathkey, value)
		} else {
			break
		}
	}
	// user trait conversion
	for i := 1; true; i++ {
		trait_key := fmt.Sprintf("%s%d", uatrk, i)
		trait_value := fmt.Sprintf("%s%d", uatrv, i)
		trait_type := fmt.Sprintf("%s%d", uatrt, i)

		if jsonParsed.Exists(trait_key) && jsonParsed.Exists(trait_value) && jsonParsed.Exists(trait_type) {
			pathkey := jsonParsed.Search(trait_key).Data().(string)
			jsonObj.Set(jsonParsed.Search(trait_type), traits, pathkey, ttype)
			jsonObj.Set(jsonParsed.Search(trait_value), traits, pathkey, value)
		} else {
			break
		}
	}
	// other values conversion
	for key, val := range lookuptable {
		if jsonParsed.Exists(key) {
			jsonObj.Set(jsonParsed.Search(key).Data().(string), val)
		}
	}

	fmt.Println(jsonObj.String())
	return []byte(jsonObj.String())
}

func isJSON(body []byte) bool {
	var js interface{}
	return json.Unmarshal(body, &js) == nil
}
