package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bitcoinschema/go-bmap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "alchema"

// Connection is a mongo client
type Connection struct {
	*mongo.Client
}

// Connect establishes a connection to the mongo db
func Connect(ctx context.Context) (*Connection, error) {
	alchemaMongoURL := os.Getenv("ALCHEMA_MONGO_URL")
	if len(alchemaMongoURL) == 0 {
		return nil, fmt.Errorf("Set ALCHEMA_MONGO_URL before running %s", alchemaMongoURL)
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(alchemaMongoURL))
	if err != nil {
		fmt.Println("Failed", err)
		return nil, err
	}

	return &Connection{client}, nil
}

// GetDocs gets a number of documents for a given collection
func (c *Connection) GetDocs(collectionName string, limit int64, skip int64, filter bson.M) ([]bmap.Tx, error) {
	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, filter, &options.FindOptions{
		Skip:  &skip,
		Limit: &limit,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())
	var txs []bmap.Tx
	for cur.Next(context.Background()) {
		// To decode into a bmap.Tx
		bmapTx := bmap.Tx{}
		err := cur.Decode(&bmapTx)
		if err != nil {
			return nil, err
		}

		txs = append(txs, bmapTx)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return txs, nil
}

// InsertOne connects and inserts the provided data into the provided collection
func (c *Connection) InsertOne(collectionName string, data bson.M) (interface{}, error) {

	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return 0, err
	}

	return res.InsertedID, nil
}

// Update connects and updates the provided data into the provided collection
// NOTE: This function can update multiple records if the filter is not restrictive
func (c *Connection) Update(collectionName string, filter interface{}, update bson.M) (interface{}, error) {

	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	log.Println("res", res)

	return res, nil
}

// UpsertOne connects and updates the provided data into the provided collection given the filter
func (c *Connection) UpsertOne(collectionName string, filter interface{}, data bson.M) (interface{}, error) {

	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	opts := options.Update().SetUpsert(true)

	update := bson.M{"$set": data}

	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, err
	}

	return res.UpsertedID, nil
}

// Upsert connects and updates the provided data into the provided collection given the filter
func (c *Connection) Upsert(collectionName string, filter interface{}, update bson.M) (interface{}, error) {

	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	opts := options.Update().SetUpsert(true)

	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, err
	}

	return res.UpsertedID, nil
}

// CountCollectionDocs returns the number of records in a given colletion
func (c *Connection) CountCollectionDocs(collectionName string, filter bson.M) (int64, error) {
	collection := c.Database(databaseName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}
