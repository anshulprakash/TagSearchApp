package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// mapping to store tags
var mapping = make(map[string][]Node)

const (
	apiKey                = "<<Insert your API key here>>"
	apiURL                = "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs"
	numberOfResults       = 10
	imagesLimitPerAPICall = 128
)

// Node is used to store information about image url and probability for the tag
type Node struct {
	ImageURL string
	Value    float64
}

// TagResp is struct for response from Clarifai's REST API
type TagResp struct {
	Status struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	} `json:"status"`
	Outputs []struct {
		ID     string `json:"id"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		CreatedAt time.Time `json:"created_at"`
		Model     struct {
			ID         string    `json:"id"`
			Name       string    `json:"name"`
			CreatedAt  time.Time `json:"created_at"`
			AppID      string    `json:"app_id"`
			OutputInfo struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				TypeExt string `json:"type_ext"`
			} `json:"output_info"`
			ModelVersion struct {
				ID        string    `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				Status    struct {
					Code        int    `json:"code"`
					Description string `json:"description"`
				} `json:"status"`
			} `json:"model_version"`
			DisplayName string `json:"display_name"`
		} `json:"model"`
		Input struct {
			ID   string `json:"id"`
			Data struct {
				Image struct {
					URL string `json:"url"`
				} `json:"image"`
			} `json:"data"`
		} `json:"input"`
		Data struct {
			Concepts []struct {
				ID    string  `json:"id"`
				Name  string  `json:"name"`
				Value float64 `json:"value"`
				AppID string  `json:"app_id"`
			} `json:"concepts"`
		} `json:"data"`
	} `json:"outputs"`
}

func main() {
	// first read the image files
	pwd, _ := os.Getwd()
	b, err := ioutil.ReadFile(pwd + "\\files\\images.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	str := string(b)            // convert content to a 'string'
	urls := strings.Fields(str) //create an array of image urls

	//Pass image urls to tagImages function to tag images and to store tags in mapping
	tagImages(urls)

	// routing requests
	http.Handle("/", http.FileServer(http.Dir("./templates")))
	http.HandleFunc("/search", search)
	http.HandleFunc("/fetchTags", fetchTags)
	http.ListenAndServe(":3000", nil)
}

/*search function looks up the tag in mapping and
returns atmost top 10 image urls in form of comma separated values to the web page*/
func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	results := ""
	searchTag := r.FormValue("searchTag") //get search value entered by user in the form

	//if the tag is present in mapping then return sorted image urls as results
	if val, ok := mapping[searchTag]; ok {
		// Tag Present
		for i, node := range val {
			if i < numberOfResults {
				results += node.ImageURL + ","
			} else {
				break
			}
		}
		results = results[:len(results)-1]
	}
	w.Write([]byte(results))
}

//fetchTags function is used to return all tags for autocomplete functionality in search field
func fetchTags(w http.ResponseWriter, r *http.Request) {

	//get all keys i.e. tags from the mapping
	keys := make([]string, 0, len(mapping))
	for k := range mapping {
		keys = append(keys, k)
	}
	allTags := strings.Join(keys, ",")
	w.Write([]byte(allTags))
}

/* tagImages function makes API calls to Clarifai's REST API and then creates a mapping for tags and
corresponding images sorted by probability */
func tagImages(urls []string) {

	// 128 images can be send in a single API call
	numberOfUrls := len(urls)
	numberOfCalls := numberOfUrls/imagesLimitPerAPICall + 1

	for i := 0; i < numberOfCalls; i++ {
		fmt.Printf("API Call Number = %d of %d\n", i+1, numberOfCalls)
		start := i * imagesLimitPerAPICall
		end := 0
		if start+imagesLimitPerAPICall < numberOfUrls {
			end = start + imagesLimitPerAPICall
		} else {
			end = numberOfUrls
		}
		// imputJSON contains image urls in form of array
		var inputJSON = `{ "inputs": [`

		for _, url := range urls[start:end] {
			inputJSON += `{
				"data": {
				  "image": {
					"url": "` + url + `"
				  }
				}
			  },`
		}
		inputJSON = inputJSON[:len(inputJSON)-1]
		inputJSON += `]}`
		var jsonStr = []byte(inputJSON)

		//API request to take images
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", "Key "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		var record TagResp

		if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
			fmt.Println(err)
		}
		// process all the tags returned for each image and store in the mapping
		for j := 0; j < len(record.Outputs); j++ {
			fmt.Printf("Storing tags for image %d \n", j+128*i+1)
			tags := record.Outputs[j].Data.Concepts
			inputURL := record.Outputs[j].Input.Data.Image.URL

			for _, v := range tags {
				var tempNode = Node{inputURL, v.Value}
				if val, ok := mapping[v.Name]; ok {
					// Tag Present
					val = append(val, tempNode)
					mapping[v.Name] = val
				} else {
					//New Tag
					mapping[v.Name] = []Node{tempNode}
				}
			}
		}
	}

	//Sort all image urls corresponding to a tag based on the probability value
	for _, element := range mapping {
		sort.Slice(element, func(i, j int) bool {
			return element[i].Value > element[j].Value
		})
	}
	fmt.Println("Clarifai's magic is done!! and now you can search at localhost:3000")
}
