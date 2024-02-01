package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type request map[string]interface{}

type main_base struct {
	Event       string                 `json:"event"`
	EventType   string                 `json:"event_type"`
	AppID       string                 `json:"app_id"`
	UserID      string                 `json:"user_id"`
	MessageID   string                 `json:"message_id"`
	PageTitle   string                 `json:"page_title"`
	PageURL     string                 `json:"page_url"`
	BrowserLang string                 `json:"browser_language"`
	ScreenSize  string                 `json:"screen_size"`
	Attributes  map[string]interface{} `json:"attributes"`
	UserTraits  map[string]interface{} `json:"traits"`
}

type msg_base struct {
	Msg string `json:"msg"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var request request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		main_base := parsed_json(request)
		go process_request(main_base)
		w.Header().Set("Content-Type", "application/json")
		webhook_res, err := send_to_webhook(main_base)
		if err != nil {
			fmt.Println("Error sending data to webhook:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := json.Marshal(webhook_res)
		// fmt.Println("responseeee: ", webhook_res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

func send_to_webhook(data main_base) (msg_base, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return msg_base{Msg: "Error occurred when sending data"}, err
	}

	resp, err := http.Post("https://webhook.site/d072748e-6f43-4191-adc1-ca02a881984b", "application/json", bytes.NewBuffer(payload)) //change web hook here
	if err != nil {
		return msg_base{Msg: "Error occurred when sending data"}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return msg_base{Msg: "Error occurred when sending data"}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return msg_base{Msg: "Data sent successfully to the webhook"}, nil
}

func parsed_json(req request) main_base {
	process_request := main_base{
		Event:       req["ev"].(string),
		EventType:   req["et"].(string),
		AppID:       req["id"].(string),
		UserID:      req["uid"].(string),
		MessageID:   req["mid"].(string),
		PageTitle:   req["t"].(string),
		PageURL:     req["p"].(string),
		BrowserLang: req["l"].(string),
		ScreenSize:  req["sc"].(string),
		Attributes:  make(map[string]interface{}),
		UserTraits:  make(map[string]interface{}),
	}

	for key, value := range req {
		if len(key) > 4 {
			switch prefix := key[:4]; prefix {
			case "atrk":
				attr_index := key[4:]
				atrk_key := fmt.Sprintf("atrv%s", attr_index)
				atrt_key := fmt.Sprintf("atrt%s", attr_index)
				// println("\n", attr_index, key, atrk_key, atrt_key)
				process_request.Attributes[value.(string)] = map[string]interface{}{
					"value": req[atrk_key].(string),
					"type":  req[atrt_key].(string),
				}
			case "uatrk":
				trait_index := key[5:]
				uatrk_key := fmt.Sprintf("uatrv%s", trait_index)
				uatrt_key := fmt.Sprintf("uatrt%s", trait_index)
				process_request.UserTraits[value.(string)] = map[string]interface{}{
					"value": req[uatrk_key].(string),
					"type":  req[uatrt_key].(string),
				}
			}
		}
	}

	return process_request
}
func process_request(main_base main_base) {
	fmt.Println("Received converted request:", main_base)
}
