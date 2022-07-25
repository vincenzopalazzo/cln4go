package plugin

import (
	"encoding/json"
	"fmt"
	"strings"

)

type GetManifestBody struct {
	Options []rpcOption `json:"options"`
	RPCMethods []rpcMethod `json:"rpcmethods"`
	Hooks []rpcHook `json:"hooks"`
	Subscriptions []string `json:"subscriptions"`
	Notifications []any `json:"notifications"`
	Dynamic bool `json:"dynamic"`
}

// func generateArray(mapData map[string]any) any {
// 	v := make([]any, 0, len(mapData))

// 	for  _, value := range mapData {
// 		v = append(v, value)
// 	 }
// 	return v
// }



type HTTPRequestBody struct {
	Reviewers []Reviewer `json:"reviewers,omitempty"`
}

type Reviewer struct {
	User User `json:"user"`
}

type User struct {
	Name string `json:"name"`
}

type Name struct {
	Name string `json:"name`
}

func addReviewers(reviewers string) []Reviewer {
	reviewersSplit := strings.Split(reviewers, ",")
	rArray := []Reviewer{}
	for _, r := range reviewersSplit {
		rArray = append(rArray, Reviewer{User: User{Name: r}})
	}
	return rArray
}

func main() {
	testReviewers := "joan,john"

	prReviewers := addReviewers(testReviewers)
	requestBody := &HTTPRequestBody{
		// Error: cannot use prReviewers (variable of type *ReviewersArray) as ReviewersArray value in struct literal
		Reviewers: prReviewers,
	}

	jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
	jsonBody, _ := json.Marshal(requestBody)
	fmt.Println(string(jsonData))
	fmt.Println(string(jsonBody))
}


Input 

Options = [
"hello" : 
{
      "name": "hello",
      "type": "string",
      "default": "World",
      "description": "What name should I call you?",
      "deprecated": false
    }, 
"greeting":  {
      "name": "greeting",
      "type": "string",
      "default": "World",
      "description": "What name should I call you?",
      "deprecated": false
    }
]

Output

{
"options": [
    {
      "name": "greeting",
      "type": "string",
      "default": "World",
      "description": "What name should I call you?",
      "deprecated": false
    }
	{
      "name": "hello",
      "type": "string",
      "default": "World",
      "description": "What name should I call you?",
      "deprecated": false
    }
  ],
}

