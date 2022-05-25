package gomek

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, schema interface{}) {
	j, err := json.Marshal(schema)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	fmt.Fprintf(w, string(j))
}
