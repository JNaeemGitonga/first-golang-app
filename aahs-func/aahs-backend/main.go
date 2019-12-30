package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	bson "go.mongodb.org/mongo-driver/bson"
)

var (
	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")

	// client
	client *mongo.Client

	// mongoURI uri string for MongoDB
	mongoURI string

	// collection global variable for collection --> if we were using multiple collections this would need to be scoped to the function executing a CRUD operation
	collection *mongo.Collection

	// newDbResult used to create a container of which we use its address for reflection see usage in handleResultSendDbResponse
	newDbResult DbResults
)

const (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// mongoURIName name of env var
	mongoURIName = "MONGO__URI"

	// getMethod get string
	getMethod = "GET"

	// postMethod post string
	postMethod = "POST"

	// putMethod put string
	putMethod = "PUT"

	// dbName name of db
	dbName = "aahsprod"

	// col name of collection
	col = "stories"
)

// Story describes the body of a Story
//without the capital letters and comments in `` this would not work
type Story struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Author  string             `json:"author,,omitempty" bson:"author,,omitempty"`
	Title   string             `json:"title,,omitempty" bson:"title,,omitempty"`
	Content string             `json:"content,,omitempty" bson:"content,,omitempty"`
	Likes   int                `json:"likes,,omitempty" bson:"likes,,omitempty"`
}

// DbResults is so that we can have one function that marshals different types
type DbResults struct {
	UpdateResult    *mongo.UpdateResult
	InsertOneResult *mongo.InsertOneResult
}

func main() {
	getMongoURIEnvVar(mongoURIName)
	if client == nil && mongoURI != "" {
		connectToMongoDB()
	}
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	method := request.HTTPMethod
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if client == nil {
		return handleError(errors.New("No connection to DB"))
	}

	switch method {
	case getMethod:
		return getStories(ctx)
	case postMethod:
		return postStory(ctx, request.Body)
	case putMethod:
		return updateStory(ctx, request.Body)
	default:
		return defaultReturn()
	}
}

func connectToMongoDB() {
	if client != nil {
		return
	}
	var connectionError error
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	client, connectionError = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	fmt.Printf("%+v", client)
	collection = client.Database(dbName).Collection(col) // <--- since I'm only using one collection this can be global
	defer cancel()

	if connectionError != nil {
		fmt.Printf("this is an eror %+v", connectionError)
	}
}

func defaultReturn() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "whatever you did, it worked!",
		StatusCode: 200,
	}, nil
}

func getMongoURIEnvVar(varName string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoURI = os.Getenv(varName)
}

func getStories(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		return handleError(err)
	}

	cursor, err := collection.Find(ctx, bson.D{})
	defer cursor.Close(ctx)

	if err != nil {
		return handleError(err)
	}

	// stories is the array of stories that will get fetched
	var stories []Story
	for cursor.Next(ctx) {
		// create a value into which the single document can be decoded
		var story Story
		err := cursor.Decode(&story)
		if err != nil {
			return handleError(err)
		}
		stories = append(stories, story)
	}

	return marshalJSONAndSend(stories)
}

func handleError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 502,
	}, err
}

/*
	handleResultSendDbResponse is used to handle multiple db responses
	from insert and update operations it can also be extended to handle
	delete operations when we are ready to add that functionality to the
	API
*/
func handleResultSendDbResponse(whichType string) (events.APIGatewayProxyResponse, error) {
	/*
		you have to use reflect to dynamically get a value from a struct as
		I am doing below. I used a struct so that I could reuse some code
		I went through this just so I could access a property using a variable that
		stored the prop name. In JS I could have done Story[whichType]
	*/
	reflectValue := reflect.ValueOf(&newDbResult)
	underlyingStruct := reflect.Indirect(reflectValue).FieldByName(whichType).Elem()
	reflectVal := underlyingStruct.Interface()

	str := fmt.Sprintf("%+v, \n", reflectVal)
	return events.APIGatewayProxyResponse{
		Body:       str,
		StatusCode: 200,
	}, nil
}

func marshalJSONAndSend(res []Story) (events.APIGatewayProxyResponse, error) {
	response, err := json.Marshal(res)
	if err != nil {
		return handleError(err)
	}
	return events.APIGatewayProxyResponse{
		Body:       string(response),
		StatusCode: 200,
	}, nil
}

func postStory(ctx context.Context, story string) (events.APIGatewayProxyResponse, error) {
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		return handleError(err)
	}

	var raw Story                   // <--- this and the folowing two lines was a big gotcha
	s := strings.NewReader(story)   // ^---> because the type of response.Body is `string` and you need it
	json.NewDecoder(s).Decode(&raw) // ^---> to be what's defined in your schema/json

	result, err := collection.InsertOne(ctx, raw)
	if err != nil {
		return handleError(err)
	}
	/*
	 the following was a big gotcha,
	 cause `result` is type *InsertOneResult
	 and you can't send that as json in your lambda response
	 I had to do some reflection to the result in order to send it.
	*/
	newDbResult = DbResults{nil, result}
	return handleResultSendDbResponse("InsertOneResult")
}

func updateStory(ctx context.Context, update string) (events.APIGatewayProxyResponse, error) {
	var raw Story          // <--- this line and the following two are important here because if we didn't do this
	data := []byte(update) // we wouldn't be able to use the incoming data as is since it is of type `string`.
	if err := json.Unmarshal(data, &raw); err != nil {
		handleError(err)
	}

	filter := bson.D{{"_id", raw.ID}} // <-- this is how we access the stored
	upDate := bson.D{{"$set", bson.D{{"likes", raw.Likes}}}}
	result, err := collection.UpdateOne(ctx, filter, upDate)
	if err != nil {
		return handleError(err)
	}

	newDbResult = DbResults{result, nil}
	return handleResultSendDbResponse("UpdateResult")
}
