---

SAM deploys GoLambda w/MongoDB
SAM! Golang…MongDB (trials + tribulations) Microservices series Part III

SAM-MongoDB-Go[pher]

---

Do you like not having a reference to build something? Well…as well known as these three technologies are, I was hard pressed to find another blog/article demonstrating a working model of them. In fact, I didn't find one. And now this one is born!
First you have to have Golang and SAM and the AWS CLI (have links for reading/installing each). This article assumes that you have some experience with Lambda functions and SAM, and MongoDB but none using them with Go. It get's long because I wanted to document the many things that I had to learn to get this thing off the ground. This was a weekend project for me and my introduction to Go. Please enjoy.
After you have everything installed you're ready to work.
First you need to cd to your Go directory. I use a Mac, so for me the directory was originally located in $HOME/go. It's important where you put this since all of your Gorutines will need to live here in the src/ directory, a tad bit more on that in a moment. So…since I like to personalize things, I decided to move it to my projects/ directory.
In order to do that and still have things work, I needed to update my $PATH. Since I use Zsh, I needed to update my .zshrc with the following lines:
export GOPATH=$HOME/porjects/go
export PATH="$PATH:$HOME/projects/go/bin"
I also needed to update my .bashrc with the export PATH="$PATH:$home/projects/go/bin since I occasionally use my terminal.
Next I made a directory for my Go Lambda as would be instructed in the Go docs using*:
$ mkdir $HOME/projects/go/src/test-go
*All of my Go programs will need to live here $HOME/projects/go/src. This is known as my Go workspace. You can read more about Go Workspaces here.
Then I run the following to create a boiler-plate Hello World Golang Lambda using the SAM CLI:
$ cd $HOME/projects/go/src/test-go
$ sam init - runtime go1.x - name test-go
When you run the above command in your terminal, you get this.
.
└── test-go
    ├── Makefile
    ├── README.md
    ├── hello-world
    │   ├── test-go
    │   ├── hello-world
    │   ├── main.go
    │   └── main_test.go
    └── template.yaml
Next we want to run:
$ make deps
$ make build
$ sam local start-api
And from a new terminal window you can run curl localhost:3000/hello or simply open that in your browser.
If all of your aws-cli credentials are in place, the first time you run $ sam local start-api you'll see something like this in your console:
Mounting HelloWorldFunction at http://127.0.0.1:3000/hello [GET]
You can now browse to the above endpoints to invoke your functions. You do not need to restart/reload SAM CLI while working on your functions, changes will be reflected instantly/automatically. You only need to restart SAM CLI if you update your AWS SAM template
2019-12-15 13:23:18  * Running on http://127.0.0.1:3000/ (Press CTRL+C to quit)
Invoking hello-world (go1.x)
2019-12-15 13:23:35 Found credentials in shared credentials file: ~/.aws/credentials
Fetching lambci/lambda:go1.x Docker container image..........................................................................................................................
Afterwards, you'll be able to invoke your lamba either using curl http://localhost:3000/hello or via the browser by visiting the same URL.
Great! You're up and running and you're developing, but now's a good time to talk about a couple of gotchas.
Now all of that was assuming that you didn't already initiatetest-go/ as a git repository. If you did and try to run the $ make deps command you may run into an error similar to the following:

$ make deps
go get -u ./...
# cd /Users/jahagitonga/projects/go/src/test-go; git submodule update --init --recursive
fatal: No url found for submodule path 'test-go' in .gitmodules
package aahs-go-back-end/aahs-func/test-go: exit status 128
SAM will not hot-reload Golang lambdas natively, which is terrible. So…in light of that, we need a Node package named supervisor to help us watch for changes. Yup, we need Node, so if you don't have it go here to find out more about it and installing it on your machine.

Now let's update our Makefile. My Makefile looks like this:



The idea for this came from Ucchishta Sivagur. His repo can be viewed here. Thanks uccmen! If you update your Makefile run $ make and $ make watch, you'll be ready to develop with hot-reloading enabled for your Golang Lambda. Sweet!
When building a lambda, you will see this error locally if you forget to declare you method in your template.yaml file:

{
  "message": "Missing Authentication Token"
}
So make sure you update your template.yaml file accordingly and don't be alarmed although the message is highly misleading. Hope the AWS crew can see about getting a better error message.
Getting MongoDb up and running!
Since this is a small project, I'm not concerned with replication or sharding; I just needed a MongoDb instance in the cloud, so I went with mLab and since it is free and they're a company owned by MongoDb. If I need to upgrade and harness the power of replication and sharding in the future, I can migrate to Atlas normal y sin problemas.
We'll need a few dependencies to make this work:
"github.com/joho/godotenv"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
"go.mongodb.org/mongo-driver/mongo/readpref"
bson "go.mongodb.org/mongo-driver/bson"
We can add the above to our import statement of our main.go file. A complete list of all the packages that I used can be found here in the repo. Let's start looking at my func main().



func main() {…}I like to keep things simple so the first thing this is going to do is get our environment variables, check to see if there is already a Mongo client and that mongoURI isn't assigned a value that's an empty string. Once we do that we connect to our MongoDb instance, or start the lambda.
So one of the first things that I noticed coming from JavaScript is that Go needs a little help when dealing with environment variables. You can make them and access them at a lower level using your system the   os package that Go gives you. But if you're used to using .env files to store these variables rather than on your system you'll need to import "github.com/joho/godotenv". Even in Node you need the dotenv package so that the Node process will gather them from your system and even your .env file(s). And to access them all you need is proces.env.YOUR_VAR_NAME. But we're not dealing with the simplicity of Node 😃! Checkout the getMongoURIEnvVar method:



func getMongoURIEnvVar() {...}

First we have to load the .env file then we can use os to get the name. Notice how load only returns an error…interesting assumption. If there wasn't an error it must have worked! Right!? 😁
Well now that we have all of that let's take a look at how we connect to our mongo instance:



Here you can see that I'm first checking to see if the client is not nil. If it isn't the program will continue and use the cached instance that is available for connection pooling. If it's nil we create a context in which our function will run and connect to the db (we have some basic logging for error handling in there too). In high level terms this context says that if the program hasn't connected in 15 seconds to stop and move on. Go has a great doc on how Google uses context It can be read here. If we pay attention to their usage we can develop some super efficient APIs. I do a decent job here but would love for the more seasoned gophers to give me some pointers on how things can be optimized here- pun intended 😆. You want to pass context around to ensure efficiency in your app and there is proper release of those resources once things have been executed within their context. Context gets deep, so please check out the two resources I listed in this paragraph for the lo'.
After we've connected or not the next step in our Gorutine is to start our handler. Take a look at it:



I try to keep it as simple as possible (KISS), right? Notice I create another context. This is the context in which our CRUD operations run in. I will pass this down to my functions and once they have returned this Gorutine will run the defer call to cancel() which will cancel the context, thus freeing up those resources. Next, I'm going to check that the client is nil if it is, which it shouldn't be, the program will send an error to the client alerting the end user that a connection couldn't be established - in so many words.
Next, I use a simple switch statement to decide what happens next. I chose to do this because the project is small and I am only using one catchall route to handle my request. If the project were bigger I wouldn't but seeing how this is an example of using Go, MongoDb and SAM…just know that everything here isn't completely up to the RESTful spec.
With that being said let's look at func getStories().



First things first, I ping the client. I do this to make sure things are on the up and up. If they're not - the program throws an error. But let's say that things are then we do a find on our collection, assess it for errors and and use the cursor's .Next() method to loop over the cursor and decode it's elements into something that our Gorutine can work with. We then append them to a stories list and return our call tofunc marshalJSONAndSend() passing it our list. The last thing done after that return executes is the deferred cancellation of cursor.Close(ctx).
But what about func marshalJSONAndSend(). Well…let's take a peek:



This little guy is going to take our list of stories, make it into a data type - []byte - that can then be converted, line 7, and sent via our API Gateway Response to the client - a string.
And that's that!
I've been using this func handleError() method but haven't shown it. There's not much to it. Take a look for your self.



func handleError(err error) (events.APIGatewayProxyResponse, error) {…}We're almost done. We have the ability to read from the db but let's create a document. We do that with our func postStory().



I follow the sam pattern that's used in the other methods so no change there but I did need to do some handle once again so that it could properly be added to my db. Once I've done that I used a variable newDbResult to create a map with of the results and the type of result. I did this to dry up some code that was used to handle the result and send the appropriate response. Before we jump to func handleResultSendResponse() let's first take a look at the Story struct. Structs remind me of interfaces in TypeScript. They are used to enumerate the fields of a piece of data. They contain names as fields and the value type associated with said fields. They also contain metadata that is of critical importance when dealing with MongoDB. Check out the Story struct:



Without that metadata our struct would not map to our JSON schema. We also needed to capitalize the field names. See what this guy had to say about capitalizing field names of structs.
https://github.com/asaskevich/govalidator/issues/187Now what about func handleResultSendResponse()?



I commented the code so…there you have it. The important thing here is reflection. Golang uses reflect to allow us to use Go in a more dynamic fashion, i.e., examine types at runtime versus compile time. Reflection deserves it's own post. So much that I found a good one here in addition to the Go docs here and here. Once we properly reflected, I was able to fmt.Sprintf a string so that we could send it back to our client, line 18 of func handleResultSendResponse(). Knowing reflection allowed me to DRY up some code for reuse in other places. Gotta love that!
Last thing we have is the update function. Let's dive in!



First let me say that I had to get used to all of the type conversion that is required to make this API run. If someone has a better way please feel free to drop me a line or two in the comments. First we declare a variable → take the update that we have and convert it to []bytes and on line four I love the way go allows us to do some shorthand assignment and use of variables in our if/else statements. Those are scoped to the if block of course, and its respective else should it have one. In fact, I liked it so much that I was willing to go against Go's linter to use this syntax with an else block that did nothing but return a function call but ended up refactoring the code.
When we look at our filter and upDate vars we see that the Go driver uses BSON. This was new to me and and the double curlies made me think of Angular 😔. Here's a great tutorial on using the Go driver.