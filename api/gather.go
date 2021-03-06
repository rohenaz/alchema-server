package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	apirouter "github.com/mrz1836/go-api-router"
	"github.com/rohenaz/alchema-server/database"
	"go.mongodb.org/mongo-driver/bson"
)

func gather(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	params := apirouter.GetParams(req)
	collection := params.GetString("collection")
	limit := params.GetInt("limit")
	skip := params.GetInt("skip")
	find := params.GetString("query")

	// decode b64 string
	decoded, err := base64.StdEncoding.DecodeString(find)

	q := bson.M{}
	err = json.Unmarshal(decoded, &q)
	if err != nil {
		log.Println(err)
		return
	}

	// DB connection
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := database.Connect(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		_ = conn.Disconnect(ctx)
	}()

	// Get matching documents
	records, err := conn.GetDocs(collection, int64(limit), int64(skip), q)
	if err != nil {
		log.Println(err)
		return
	}

	apirouter.ReturnResponse(w, req, http.StatusOK, records)
}
