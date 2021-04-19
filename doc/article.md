# Digitalocean Golang file uploads API


### Introduction

The ability of a user to upload their personal files when using an application is often considered as a needed feature. However, when using a GraphQL API, this feature could become a challenge to implement, especially with GraphQL’s single source of truth design in your client application.

In this article, you would learn how file uploads functionality can be built out in a GraphQL API implemented in Golang. To make the tutorial easy to follow through, it has been broken down into smaller steps;                                                                                                                                                                                      

## Prerequisites

To get the best out of this article, you would need the following;



-  Basic knowledge of Golang. If you are new to Golang, this article provides an explanation of Golang, including how to configure your local machine for developing with Golang.
-   An understanding of Graph Query Language ( GraphQL ). You can learn more about GraphQL by following the 
-   An active Digitalocean account, as several Digitalocean resources are used within this article.


## Step 1 — Bootstrapping a Golang GraphQL API

You would be using the Gqlgen package for creating a GraphQL API implemented in Golang.

Execute the command below from your terminal to create a new Golang project; 

```command 
 go mod init
```

Next, install the Gqlgen package; 
```command
 go get https://github.com/vektah/gqlgen
```

Then generate a boilerplate GraphQL project having files needed for a GraphQL API;

```command
 gqlgen init 
```


			

With the boilerplate application generated above, you would have a default todo application with a basic user schema structure.


### Defining Application GraphQL Schema

By default, the `gqlgen init` command would generate a boilerplate application having a User linked to a todo as the default schema structure in the schema.graphqls file. 

Replace the boilerplate code in the schema.graphqls file with the schema below;


``` [label schema.graphls]

 type Query {
   getUser: User! 
}

 type FileUpload {
   file: Upload!
   userId: String
 }

 type Mutations { 
  createUser: newUser!
 }

 input newUser {
   fullname: String!
   email: String!
   dateCreated: String!
   img_uri: String
 }

 type User { 
   fullname: String!
   email: String!
   dateCreated: String!
   img_uri: String
 }
```


The code snippet above contains a schema with three types; the Upload and User types which are known as Object types in the GraphQL Schema Definition Language and the Mutation and Query types respectively. 

**Note**: The Upload scalar type is automatically defined by Gqlgen and it contains the properties of a file. To use it, you only need to declare it at the top of the schema file, as done in the file above.

You have defined the structure of the data within this application, the next step is to generate the query and mutation resolvers functions for the schema above. 


## Step 2 — Generating Application Resolvers

The Gqlgen package being used is based on a schema first approach. A time-saving feature of Gqlgen is its ability to generate your application’s resolvers based on your defined schema file. With this feature, you do not need to manually write the resolver boilerplate code, all you need to do is to focus on the actual implementation of the defined resolvers.

To utilize the code generation, run `gqlgen generate` from a terminal within your project and observe the code below being added to the `schema.resolvers.go` file.


## Step 3 — Provisioning and Using a Managed Database Instance on DigitalOcean 

Although the application would not store images directly in a database, it still needs a database to insert each user‘s record. The stored record would then contain links to the uploaded files.

A user’s record would consist of a **Fullname**, **email**, **dateCreated,** and an **img_uri** field of String data type. The **img_uri** field would contain the URL pointing to an image file uploaded by a user through this GraphQL API and stored within a bucket on Digitalocean spaces.

Using your Digitalocean dashboard, navigate to the Databases section of the console to create a new database cluster. By default, PostgreSQL would be the selected database to run within this cluster. Leave all other settings at their default values and proceed to create this cluster using the button at the bottom. 

After the cluster has been created, the connection details of the cluster would be displayed. Create a `.env` file within the GraphQL-API project directory to securely store the cluster values in the following format;

  


```[label .env]

 DB_PASSWORD=<PASSWORD>
 DB_PORT=<PORT>
 DB_NAME=<DATABASE>
 DB_ADDR=<HOST>
 DB_USER=<USERNAME>
```


With the connection details securely stored in the .env file above, the next step would be to connect to the database cluster through our backend application.

Create a db.go file within the `graph` directory and add the code below which establishes a database connection, into the file;


```[label server.go]

package db

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/vickywane/api/graph/model"
	"os"
)

func createSchema(db *pg.DB) error {
	for _, models := range []interface{}{(*model.User)(nil), (*model.User)(nil)}{
		if err := db.CreateTable(models, &orm.CreateTableOptions{
			IfNotExists: true
		}); err != nil {
			panic(err)
		}
	}

	return nil
}

func Connect() *pg.DB {
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	DB_ADDR := os.Getenv("DB_ADDR")
	DB_USER := os.Getenv("DB_USER")

	connStr := fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v?sslmode=require",
		DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_NAME )

	opt, err := pg.ParseURL(connStr); if err != nil {
  	  panic(err)
      }

	db := pg.Connect(opt)

	if schemaErr := createSchema(db); schemaErr != nil {
		panic(schemaErr)
	}

	if _, DBStatus := db.Exec("SELECT 1"); DBStatus != nil {
		panic("PostgreSQL is down")
	}

	return db
}
```


Through the code snippet above, the application can establish a connection with the managed instance on DigitalOcean to access the PostgreSQL database through the following steps;



*   Starting from the exported `Connect` function, the stored database instance credentials from our `.env` file are retrieved, then, a string is formatted with the credentials to be used as a connection URI.
*   Next, the created connection string is parsed and passed into the `pg.connect` method as an argument to open a connection.
*   As the last step, database tables are created using the models that would be generated by `gqlgen generate` command later on. 

Next, you need to add this package to the main application so the database connection would be established at startup and also be available in the `Resolver` struct.

Open the server.go file in your preferred code editor. You need to modify the `server.go` file with the code snippet below to utilize the previously created `db` package immediately after the application is started.


```[label db/db.go]
 package main

import (
  "log"
  "net/http"
  "os"

  "github.com/vickywane/api/graph/db"
  "github.com/99designs/gqlgen/graphql/handler"
  "github.com/99designs/gqlgen/graphql/playground"
  "github.com/vickywane/api/graph"
  "github.com/vickywane/api/graph/generated"
)

const defaultPort = "8080"

func main() {
  port := os.Getenv("PORT")
  if port == "" {
     port = defaultPort
  }

  Database := db.Connect()
  srv := handler.NewDefaultServer(
    generated.NewExecutableSchema(generated.Config{
       Resolvers: &graph.Resolver{
    		DB: Database,
 	 }
    }))

  http.Handle("/", playground.Handler("GraphQL playground", "/query"))
  http.Handle("/query", srv)

  log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
```


Within the code snippet above, you expressed the Connect function from the DB package in the `Database` variable then you passed in the connected database client into the graph resolver.

Lastly, you need to specify the data type of the DB field you added in the Resolver struct above.

To achieve this, open the `resolver.go` file and modify the Resolver struct to have a DB field with a `go-pg` pointer as its type as shown below;


``` [label resolver.go]
package resolvers

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"sync"

	"github.com/vickywane/event-server/graph/model"
)

type Resolver struct {
	DB *pg.DB
}
```


Now a database connection would be established each time the entry `server.go` file is run and the `go-pg` package can be used as an ORM to perform operations on the database from the resolver functions. 


## Step 4 — Implementing Generated Resolvers

In GraphQL, a resolver is a function that resolves the value for a field in the defined schema. Utilizing Gqlgen’s code generation feature, you can generate the boilerplate functions for the defined resolvers using this feature, then focus on implementing how the fields are resolved for the Query and Mutation resolvers next.


#### Mutation Resolvers

Going through the `schema.graphqls` file, there are only two mutation resolvers generated. One with the purpose of handling the user creation, while the other to handle the profile image uploads.

Modify the `CreateUser` mutation with the code snippet below to insert a new row containing the user details input into the database


```[label schema.resolver.go]
package graph

import (
  "context"
  "fmt"
  "time"

  "github.com/satori/go.uuid"
  "github.com/vickywane/api/graph/generated"
  "github.com/vickywane/api/graph/model"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
  user := model.User{
     ID:          fmt.Sprintf("%v", uuid.NewV4()),
     FullName:    input.FullName,
     Password:    input.Password,
     Email:       input.Email,
     ImgURI:      "https://bit.ly/3mCSn2i",
     DateCreated: time.Now().Format("01-02-2006"),
  }

  if err := r.DB.Insert(&user); err != nil {
     return nil, fmt.Errorf("error inserting user: %v", err)
  }

  return &user, nil
}
```


Going through the CreateUser mutation in the code snippet above, you would observe two things about the user rows inserted;



*   Each row inserted is given a unique UUID formatted as a string
*   Each row is given a placeholder image to be used until an actual profile image uploaded by the user

At this point, you have the `UploadProfileImage` mutation resolver function left to implement, but before you implement this function, you need to implement the query resolver first. This is because  each upload is linked to a specific user, hence the need to retrieve the ID of a specific user before uploading an image.


#### Query Resolver

As defined in the `schema.graphqls` file, one query resolver was generated for the purpose of retrieving all created users. 

Modify the generated `Users` query resolver with the code snippet below to query all user rows within the database.


```[label schema.graphqls]
package graph

import (
  "context"
  "fmt"
  "time"

  "github.com/satori/go.uuid"
  "github.com/vickywane/api/graph/generated"
  "github.com/vickywane/api/graph/model"
)

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
  var users []*model.User

  r.DB.Model(&users).Select()

  return users, nil  
}
```


From the code snippet above, you are able to fetch all users by using go-pg’s `select` method on the `User` model without a where or limit clause attached.

 

<p id="gdcalert1" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image1.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert2">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image1.png "image_tooltip")


At this point, you have now implemented both the `CreateUser` mutation and the `User` query. All is in place for you to move on to uploading images for a user to a bucket within Digitalocean spaces.


### Uploading Images To Digitalocean Spaces

To begin, navigate to the Spaces section of your DigitalOcean console where you would create a new bucket for storing the uploaded files from your backend application.

Click the **Create New Space** button, leave other settings at their default values and specify a unique name for the new space as shown below before creating your new space;



<p id="gdcalert2" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image2.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert3">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image2.png "image_tooltip")


After the new space has been created, navigate to the settings tab and copy the space’s endpoint into the GraphQL project environment variables.


```[label .env]
SPACE_ENDPOINT=<BUCKET_ENDPOINT>
```


Next, using t[his guide](https://docs.digitalocean.com/products/spaces/how-to/manage-access/#access-keys) within the Digitalocean Spaces documentation that explains the process of creating secret and access keys, create a secret and access key for the backend application. After creating them, copy the values and store them in your backend application’s `.env` file in the format below;


```[label .env]
ACCESS_KEY=<SPACE_ACCESS_KEY>
SECRET_KEY=<SPACE_SECRET_KEY>
```


One way to perform operations on your Digitalocean Space is through the use of supported AWS SDKs as they are quite compatible. The Digitalocean Spaces [documentation](https://docs.digitalocean.com/products/spaces/) provides a list of operations you can perform on the Spaces API using AWS SDK.

To implement the file uploads, modify the `UploadProfileImage` mutation function within the `Schema.resolvers.go` file with the code below which uploads an image from the resolver function into our Digitalocean Spaces bucket.


```[label schema.resolvers.go]
package graph

import (
   "bytes"
   "context"
   "fmt"
   "os"

   "github.com/aws/aws-sdk-go/aws"
   "github.com/aws/aws-sdk-go/aws/credentials"
   "github.com/aws/aws-sdk-go/aws/session"
   "github.com/aws/aws-sdk-go/service/s3"
   "io"

   "github.com/vickywane/api/graph/generated"
   "github.com/vickywane/api/graph/model"
)

func (r *mutationResolver) UploadProfileImage(ctx context.Context, input model.ProfileImage) (bool, error) {

   SpaceName := os.Getenv("DO_SPACE_NAME")
   SpaceRegion := os.Getenv("DO_SPACE_REGION")
   key := os.Getenv("ACCESS_KEY")
   secret := os.Getenv("ACCESS_SECRET")

   user, userErr := r.GetUserField("ID", *input.UserID)

   if userErr != nil {
       return false, fmt.Errorf("error getting user: %v", userErr)
   }

   s3Config := &aws.Config{
       Credentials: credentials.NewStaticCredentials(key, secret, ""),
       Endpoint:    aws.String(fmt.Sprintf("https://%v.digitaloceanspaces.com", SpaceRegion)),
       Region:      aws.String(SpaceRegion),
   }

   newSession := session.New(s3Config)
   s3Client := s3.New(newSession)

   stream, readErr := io.ReadAll(input.File.File)
   if readErr != nil {
       fmt.Printf("error from file %v", readErr)
   }

   fileErr := os.WriteFile("image.png", stream, 0644)
   if fileErr != nil {
       fmt.Printf("file err %v", fileErr)
   }

   file, openErr := os.Open("image.png")
   if openErr != nil {
       return false, fmt.Errorf("Error opening temporary file: %v", openErr)
   }

   defer file.Close()

   buffer := make([]byte, input.File.Size)

   file.Read(buffer)

   fileBytes := bytes.NewReader(buffer)

   object := s3.PutObjectInput{
       Bucket: aws.String(SpaceName),
       Key:    aws.String(fmt.Sprintf("%v-%v", *input.UserID, input.File.Filename)),
       Body:   fileBytes,
       ACL:    aws.String("public-read"),
   }

   if _, uploadErr := s3Client.PutObject(&object); uploadErr != nil {
       return false, fmt.Errorf("error uploading file: %v", uploadErr)
   }

   os.Remove("image.png")
   user.ImgURI = fmt.Sprintf("https://%v.%v.digitaloceanspaces.com/%v-%v", SpaceName, SpaceRegion, *input.UserID, input.File.Filename)

   if _, err := r.UpdateUser(user); err != nil {
       return false, fmt.Errorf("Err updating user: %v", err)
   }

   return true, nil
}
```


Going through the entire function above, you would observe the following being performed in the following order;



*   First, using the helper function you wrote in the `resolver.go` file, the user row having the UserID argument as an ID is queried from the database to confirm that you are trying to upload a file for an actual user.
*   Next, you configured the SDK client for Digitalocean spaces using an access key and secret key credentials obtained from the Digitalocean console.
*   Next, using the `ReadAll` method from the `io` package, you read the entire content of the file added to the HTTP request sent to the GraphQL API, then a temporary file was created to dump this content into.
*   Next, you added the `PutObjectInput` struct fields having the `Bucket` field as the name of the Space on DigitalOcean, the `Key` field as the name of the file being uploaded, the `Body` field as the temporary file you created, and lastly the ACL ( Access Control Lock) field to set the permission type on the file.

	Note: you want the uploaded files to be public and viewed by everyone with a link, hence you used the `public-read` canned ACL type.



*   Lastly, after the file is uploaded, the temporary file is deleted, and a link to the uploaded file is formatted together using the Spaces endpoint and filename, then the user’s ImgUri field in the row is updated to contain the formatted link.

It is important to note that the reason why you are using a temporary file to store the file to be uploaded is that when using the AWS v1 SDK for Golang, the `body` property within the `PutObjectInput` struct has a `Reader.seek` type. You should put this into consideration if you expect your users to upload files with huge sizes.

To test the new mutation resolver, execute the command below to make an HTTP request to the GraphQL API using cURL, adding an image into the request form body. 


```


curl localhost:8080/query  -F operations='{ "query": "mutation uploadProfileImage($image: Upload! $userId : String!) { uploadProfileImage(input: { file: $image  userId : $userId}) }", "variables": { "image": null, "userId" : "121212" } }' -F map='{ "0": ["variables.image"] }'  -F 0=@sample.jpeg
```



#### After the file has been uploaded, the following boolean status would be printed out in the terminal as the request-response, indicating that the file upload was successful. 


```
> {"data": { "uploadProfileImage": true }}
```



#### Going through your created bucket within the Spaces section of the Digitalocean console, you would find the image recently uploaded from your terminal.



<p id="gdcalert3" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image3.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert4">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image3.png "image_tooltip")



#### Also, if you query the user’s data after a successful file upload, you would observe that the img_uri field returned in the user’s data points to the file recently uploaded to your bucket.



<p id="gdcalert4" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image4.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert5">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image4.png "image_tooltip")


At this point, you have a functional backend application exposing a GraphQL with mutation resolvers that you can use to insert a new user record into a connected Postgresql database and also upload an image for the new user.

you can move a step further to deploy this application to the Digitalocean App platform. Using the App Platform’s support for Golang, a deployment only requires a minimal configuration.


### Deploying GraphQL API to App Platform

[App platform](https://www.digitalocean.com/products/app-platform/) is a Digitalocean service product that makes it much easier to build, deploy, and even scale your applications. App platform supports a variety of languages and within this article, you would utilize the support for applications written in Go and stored within GitHub.

To begin shipping your code, create a local git repository and add your latest changes into the local repository by running the commands below one at a time from your terminal;


```
 # create a git repository
 git init

 # add latest changes in a git repository 
 git add . && git commit -m "feat: implemented upload functionality in GraphQL API"
```


Next, create a public repository on GitHub using the [Github guide](https://docs.github.com/en/github/getting-started-with-github/create-a-repo) and push the local source code into your new repository.

From your Digitalocean dashboard, navigate to the Apps section and select GitHub as the source to connect your Digitalocean account to GitHub. After the integration, select the newly created repository above from the Repository dropdown. 

In the next configuration page, define the environment variables for the application as defined in your local `.env` file as shown below;



<p id="gdcalert5" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image5.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert6">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image5.png "image_tooltip")


Leaving other settings at their defaults, click the **Next** button to move to the next page where you would give this deployment a unique name, then navigate to the remaining pages to finalize the deployment and build the app.


### Step 6 — Testing The Deployed GraphQL Endpoint

At this point, the application has been fully deployed to DigitalOcean, with a healthy running status similar to the one shown in the image below;



<p id="gdcalert6" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image6.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert7">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image6.png "image_tooltip")


Take note of the endpoint URL of your deployed application placed below the application name. you would use this endpoint to test the upload feature implemented in the deployed GraphQL API with Postman as an API testing tool.

**Note**: _If you do not have the Postman Desktop App installed on your local machine, you can make use of the Postman Web Client within your browser._

From your Postman collection, create a new POST request with a form-data body having the following keys; 

Note: You should replace the USER_ID placeholder with the ID of a user you created using the `CreateUser` mutation. You would also need to change the file field below to a file type before you can select a file on your machine.



*   operations: [{"key":"operations","value":"{\"query\": \"mutation uploadImage($userID: String! $image: Upload!) {\\n uploadProfileImage(input: { userId: $userID file: $image }
*   map: {"file": ["variables.image"], "userId": ["variables.userID"]}
*   userId: &lt;USER_ID>
*   file: &lt;LOCAL_FILE>



<p id="gdcalert7" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image7.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert8">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image7.png "image_tooltip")


Hit the Send button to send the POST request, then reload your Digitalocean Space bucket to see your newly uploaded file.


### Conclusion

By reading this article, you have learned about Digitalocean Spaces, and how you can perform file uploads to a created bucket on Digitalocean spaces using the AWS SDK for Golang from a mutation resolver in a GraphQL application.

If you would like to learn more about using Digitalocean Spaces, we encourage you to have a look at its [documentation](https://docs.digitalocean.com/products/spaces/), as it contains an explanation of all aspects of Digitalocean Spaces.

 
## Digitalocean Golang file uploads API

**Introduction**

The ability of a user to upload their personal files when using an application is often considered as a needed feature. However, when using a GraphQL API, this feature could become a challenge to implement, especially with GraphQL’s single source of truth design.

In this article, you would learn how file uploads functionality can be built out in a GraphQL API implemented in Golang. To make the tutorial easy to follow through, it has been broken down into smaller sections. You can either follow them in the order which they appear below or skip to what part interests you most;



*   Bootstrapping a Golang GraphQL API
*   Defining Application GraphQL Schema
*   Implementing Schema resolvers
*   Digitalocean Spaces API for file Storage
*   Deploying Golang GraphQL application to the App platform.                                                           

**                                                                                                                                   **

**Prerequisites**

To get the best out of this article, you would need the following;



*   Basic knowledge of Golang. If you are new to Golang, this article provides an explanation of Golang, including how to configure your local machine for developing with Golang.
*   An understanding of Graph Query Language ( GraphQL ). 
*   An active Digitalocean account, as several Digitalocean resources are used within this article.


### Step 1 — Bootstrapping a Golang GraphQL API

You would be using the Gqlgen package for creating a GraphQL API implemented in Golang.

Execute the command below from your terminal to create a new Golang project, install Gqlgen and lastly generate a boilerplate GraphQL project having files needed for a GraphQL API;


```
 # Create a golang project
 go mod init

 # install gqlgen
 go get https://github.com/vektah/gqlgen

 # create a boilerplate application 
 gqlgen init 
```


			

With the boilerplate application generated above, you would have a default todo application with a basic user schema structure.


### Defining Application GraphQL Schema

By default, the `gqlgen init` command would generate a boilerplate application having a User linked to a todo as the default schema structure in the schema.graphqls file. 

Replace the boilerplate code in the schema.graphqls file with the schema below;


```
 type Query {
   getUser: User! 
}

 type FileUpload {
   file: Upload!
   userId: String
 }

 type Mutations { 
  createUser: newUser!
 }

 input newUser {
   fullname: String!
   email: String!
   dateCreated: String!
   img_uri: String
 }

 type User { 
   fullname: String!
   email: String!
   dateCreated: String!
   img_uri: String
 }
```


The code snippet above contains a schema with three types; the Upload and User types which are known as Object types in the GraphQL Schema Definition Language and the Mutation and Query types respectively. 

**Note**: The Upload scalar type is automatically defined by Gqlgen and it contains the properties of a file. To use it, you only need to declare it at the top of the schema file, as done in the file above.

You have defined the structure of the data within this application, the next step is to generate the query and mutation resolvers functions for the schema above. 


### Step 2 — Generating Application Resolvers

The Gqlgen package being used is based on a schema first approach. A time-saving feature of Gqlgen is its ability to generate your application’s resolvers based on your defined schema file. With this feature, you do not need to manually write the resolver boilerplate code, all you need to do is to focus on the actual implementation of the defined resolvers.

To utilize the code generation, run `gqlgen generate` from a terminal within your project and observe the code below being added to the `schema.resolvers.go` file.


### Step 3 — Provisioning and Using a Managed Database Instance on DigitalOcean 

Although the application would not store images directly in a database, it still needs a database to insert each user‘s record. The stored record would then contain links to the uploaded files.

A user’s record would consist of a **Fullname**, **email**, **dateCreated,** and an **img_uri** field of String data type. The **img_uri** field would contain the URL pointing to an image file uploaded by a user through this GraphQL API and stored within a bucket on Digitalocean spaces.

Using your Digitalocean dashboard, navigate to the Databases section of the console to create a new database cluster. By default, PostgreSQL would be the selected database to run within this cluster. Leave all other settings at their default values and proceed to create this cluster using the button at the bottom. 

After the cluster has been created, the connection details of the cluster would be displayed. Create a `.env` file within the GraphQL-API project directory to securely store the cluster values in the following format;

  


```
 DB_PASSWORD=<PASSWORD>
 DB_PORT=<PORT>
 DB_NAME=<DATABASE>
 DB_ADDR=<HOST>
 DB_USER=<USERNAME>
```


With the connection details securely stored in the .env file above, the next step would be to connect to the database cluster through our backend application.

Create a db.go file within the `graph` directory and add the code below which establishes a database connection, into the file;


```
package db

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/vickywane/api/graph/model"
	"os"
)

func createSchema(db *pg.DB) error {
	for _, models := range []interface{}{(*model.User)(nil), (*model.User)(nil)}{
		if err := db.CreateTable(models, &orm.CreateTableOptions{
			IfNotExists: true
		}); err != nil {
			panic(err)
		}
	}

	return nil
}

func Connect() *pg.DB {
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	DB_ADDR := os.Getenv("DB_ADDR")
	DB_USER := os.Getenv("DB_USER")

	connStr := fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v?sslmode=require",
		DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_NAME )

	opt, err := pg.ParseURL(connStr); if err != nil {
  	  panic(err)
      }

	db := pg.Connect(opt)

	if schemaErr := createSchema(db); schemaErr != nil {
		panic(schemaErr)
	}

	if _, DBStatus := db.Exec("SELECT 1"); DBStatus != nil {
		panic("PostgreSQL is down")
	}

	return db
}
```


Through the code snippet above, the application can establish a connection with the managed instance on DigitalOcean to access the PostgreSQL database through the following steps;



*   Starting from the exported `Connect` function, the stored database instance credentials from our `.env` file are retrieved, then, a string is formatted with the credentials to be used as a connection URI.
*   Next, the created connection string is parsed and passed into the `pg.connect` method as an argument to open a connection.
*   As the last step, database tables are created using the models that would be generated by `gqlgen generate` command later on. 

Next, you need to add this package to the main application so the database connection would be established at startup and also be available in the `Resolver` struct.

Open the server.go file in your preferred code editor. You need to modify the `server.go` file with the code snippet below to utilize the previously created `db` package immediately after the application is started.


```
 package main

import (
  "log"
  "net/http"
  "os"

  "github.com/vickywane/api/graph/db"
  "github.com/99designs/gqlgen/graphql/handler"
  "github.com/99designs/gqlgen/graphql/playground"
  "github.com/vickywane/api/graph"
  "github.com/vickywane/api/graph/generated"
)

const defaultPort = "8080"

func main() {
  port := os.Getenv("PORT")
  if port == "" {
     port = defaultPort
  }

  Database := db.Connect()
  srv := handler.NewDefaultServer(
    generated.NewExecutableSchema(generated.Config{
       Resolvers: &graph.Resolver{
    		DB: Database,
 	 }
    }))

  http.Handle("/", playground.Handler("GraphQL playground", "/query"))
  http.Handle("/query", srv)

  log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
```


Within the code snippet above, you expressed the Connect function from the DB package in the `Database` variable then you passed in the connected database client into the graph resolver.

Lastly, you need to specify the data type of the DB field you added in the Resolver struct above.

To achieve this, open the `resolver.js` file and modify the Resolver struct to have a DB field with a `go-pg` pointer as its type as shown below;


```
package resolvers

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"sync"

	"github.com/vickywane/event-server/graph/model"
)

type Resolver struct {
	DB *pg.DB
}
```


Now a database connection would be established each time the entry `server.go` file is run and the `go-pg` package can be used as an ORM to perform operations on the database from the resolver functions. 


### Step 4 — Implementing Generated Resolvers

In GraphQL, a resolver is a function that resolves the value for a field in the defined schema. Utilizing Gqlgen’s code generation feature, you can generate the boilerplate functions for the defined resolvers using this feature, then focus on implementing how the fields are resolved for the Query and Mutation resolvers next.


#### Mutation Resolvers

Going through the `schema.graphqls` file, there are only two mutation resolvers generated. One with the purpose of handling the user creation, while the other to handle the profile image uploads.

Modify the `CreateUser` mutation with the code snippet below to insert a new row containing the user details input into the database


```
package graph

import (
  "context"
  "fmt"
  "time"

  "github.com/satori/go.uuid"
  "github.com/vickywane/api/graph/generated"
  "github.com/vickywane/api/graph/model"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
  user := model.User{
     ID:          fmt.Sprintf("%v", uuid.NewV4()),
     FullName:    input.FullName,
     Password:    input.Password,
     Email:       input.Email,
     ImgURI:      "https://bit.ly/3mCSn2i",
     DateCreated: time.Now().Format("01-02-2006"),
  }

  if err := r.DB.Insert(&user); err != nil {
     return nil, fmt.Errorf("error inserting user: %v", err)
  }

  return &user, nil
}
```


Going through the CreateUser mutation in the code snippet above, you would observe two things about the user rows inserted;



*   Each row inserted is given a unique UUID formatted as a string
*   Each row is given a placeholder image to be used until an actual profile image uploaded by the user

At this point, you have the `UploadProfileImage` mutation resolver function left to implement, but before you implement this function, you need to implement the query resolver first. This is because  each upload is linked to a specific user, hence the need to retrieve the ID of a specific user before uploading an image.


#### Query Resolver

As defined in the `schema.graphqls` file, one query resolver was generated for the purpose of retrieving all created users. 

Modify the generated `Users` query resolver with the code snippet below to query all user rows within the database.


```
package graph

import (
  "context"
  "fmt"
  "time"

  "github.com/satori/go.uuid"
  "github.com/vickywane/api/graph/generated"
  "github.com/vickywane/api/graph/model"
)

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
  var users []*model.User

  r.DB.Model(&users).Select()

  return users, nil  
}
```


From the code snippet above, you are able to fetch all users by using go-pg’s `select` method on the `User` model without a where or limit clause attached.

 

<p id="gdcalert1" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image1.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert2">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image1.png "image_tooltip")


At this point, you have now implemented both the `CreateUser` mutation and the `User` query. All is in place for you to move on to uploading images for a user to a bucket within Digitalocean spaces.


### Uploading Images To Digitalocean Spaces

To begin, navigate to the Spaces section of your DigitalOcean console where you would create a new bucket for storing the uploaded files from your backend application.

Click the **Create New Space** button, leave other settings at their default values and specify a unique name for the new space as shown below before creating your new space;



<p id="gdcalert2" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image2.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert3">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image2.png "image_tooltip")


After the new space has been created, navigate to the settings tab and copy the space’s endpoint into the GraphQL project environment variables.


```
SPACE_ENDPOINT=<BUCKET_ENDPOINT>
```


Next, using t[his guide](https://docs.digitalocean.com/products/spaces/how-to/manage-access/#access-keys) within the Digitalocean Spaces documentation that explains the process of creating secret and access keys, create a secret and access key for the backend application. After creating them, copy the values and store them in your backend application’s `.env` file in the format below;


```
ACCESS_KEY=<SPACE_ACCESS_KEY>
SECRET_KEY=<SPACE_SECRET_KEY>
```


One way to perform operations on your Digitalocean Space is through the use of supported AWS SDKs as they are quite compatible. The Digitalocean Spaces [documentation](https://docs.digitalocean.com/products/spaces/) provides a list of operations you can perform on the Spaces API using AWS SDK.

To implement the file uploads, modify the `UploadProfileImage` mutation function within the `Schema.resolvers.go` file with the code below which uploads an image from the resolver function into our Digitalocean Spaces bucket.


```
package graph

import (
   "bytes"
   "context"
   "fmt"
   "os"

   "github.com/aws/aws-sdk-go/aws"
   "github.com/aws/aws-sdk-go/aws/credentials"
   "github.com/aws/aws-sdk-go/aws/session"
   "github.com/aws/aws-sdk-go/service/s3"
   "io"

   "github.com/vickywane/api/graph/generated"
   "github.com/vickywane/api/graph/model"
)

func (r *mutationResolver) UploadProfileImage(ctx context.Context, input model.ProfileImage) (bool, error) {

   SpaceName := os.Getenv("DO_SPACE_NAME")
   SpaceRegion := os.Getenv("DO_SPACE_REGION")
   key := os.Getenv("ACCESS_KEY")
   secret := os.Getenv("ACCESS_SECRET")

   user, userErr := r.GetUserField("ID", *input.UserID)

   if userErr != nil {
       return false, fmt.Errorf("error getting user: %v", userErr)
   }

   s3Config := &aws.Config{
       Credentials: credentials.NewStaticCredentials(key, secret, ""),
       Endpoint:    aws.String(fmt.Sprintf("https://%v.digitaloceanspaces.com", SpaceRegion)),
       Region:      aws.String(SpaceRegion),
   }

   newSession := session.New(s3Config)
   s3Client := s3.New(newSession)

   stream, readErr := io.ReadAll(input.File.File)
   if readErr != nil {
       fmt.Printf("error from file %v", readErr)
   }

   fileErr := os.WriteFile("image.png", stream, 0644)
   if fileErr != nil {
       fmt.Printf("file err %v", fileErr)
   }

   file, openErr := os.Open("image.png")
   if openErr != nil {
       return false, fmt.Errorf("Error opening temporary file: %v", openErr)
   }

   defer file.Close()

   buffer := make([]byte, input.File.Size)

   file.Read(buffer)

   fileBytes := bytes.NewReader(buffer)

   object := s3.PutObjectInput{
       Bucket: aws.String(SpaceName),
       Key:    aws.String(fmt.Sprintf("%v-%v", *input.UserID, input.File.Filename)),
       Body:   fileBytes,
       ACL:    aws.String("public-read"),
   }

   if _, uploadErr := s3Client.PutObject(&object); uploadErr != nil {
       return false, fmt.Errorf("error uploading file: %v", uploadErr)
   }

   os.Remove("image.png")
   user.ImgURI = fmt.Sprintf("https://%v.%v.digitaloceanspaces.com/%v-%v", SpaceName, SpaceRegion, *input.UserID, input.File.Filename)

   if _, err := r.UpdateUser(user); err != nil {
       return false, fmt.Errorf("Err updating user: %v", err)
   }

   return true, nil
}
```


Going through the entire function above, you would observe the following being performed in the following order;



*   First, using the helper function you wrote in the `resolver.go` file, the user row having the UserID argument as an ID is queried from the database to confirm that you are trying to upload a file for an actual user.
*   Next, you configured the SDK client for Digitalocean spaces using an access key and secret key credentials obtained from the Digitalocean console.
*   Next, using the `ReadAll` method from the `io` package, you read the entire content of the file added to the HTTP request sent to the GraphQL API, then a temporary file was created to dump this content into.
*   Next, you added the `PutObjectInput` struct fields having the `Bucket` field as the name of the Space on DigitalOcean, the `Key` field as the name of the file being uploaded, the `Body` field as the temporary file you created, and lastly the ACL ( Access Control Lock) field to set the permission type on the file.

	Note: you want the uploaded files to be public and viewed by everyone with a link, hence you used the `public-read` canned ACL type.



*   Lastly, after the file is uploaded, the temporary file is deleted, and a link to the uploaded file is formatted together using the Spaces endpoint and filename, then the user’s ImgUri field in the row is updated to contain the formatted link.

It is important to note that the reason why you are using a temporary file to store the file to be uploaded is that when using the AWS v1 SDK for Golang, the `body` property within the `PutObjectInput` struct has a `Reader.seek` type. You should put this into consideration if you expect your users to upload files with huge sizes.

To test the new mutation resolver, execute the command below to make an HTTP request to the GraphQL API using cURL, adding an image into the request form body. 


```


curl localhost:8080/query  -F operations='{ "query": "mutation uploadProfileImage($image: Upload! $userId : String!) { uploadProfileImage(input: { file: $image  userId : $userId}) }", "variables": { "image": null, "userId" : "121212" } }' -F map='{ "0": ["variables.image"] }'  -F 0=@sample.jpeg
```



#### After the file has been uploaded, the following boolean status would be printed out in the terminal as the request-response, indicating that the file upload was successful. 


```
> {"data": { "uploadProfileImage": true }}
```



#### Going through your created bucket within the Spaces section of the Digitalocean console, you would find the image recently uploaded from your terminal.



<p id="gdcalert3" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image3.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert4">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image3.png "image_tooltip")



#### Also, if you query the user’s data after a successful file upload, you would observe that the img_uri field returned in the user’s data points to the file recently uploaded to your bucket.



<p id="gdcalert4" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image4.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert5">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image4.png "image_tooltip")


At this point, you have a functional backend application exposing a GraphQL with mutation resolvers that you can use to insert a new user record into a connected Postgresql database and also upload an image for the new user.

you can move a step further to deploy this application to the Digitalocean App platform. Using the App Platform’s support for Golang, a deployment only requires a minimal configuration.


### Deploying GraphQL API to App Platform

[App platform](https://www.digitalocean.com/products/app-platform/) is a Digitalocean service product that makes it much easier to build, deploy, and even scale your applications. App platform supports a variety of languages and within this article, you would utilize the support for applications written in Go and stored within GitHub.

To begin shipping your code, create a local git repository and add your latest changes into the local repository by running the commands below one at a time from your terminal;


```
 # create a git repository
 git init

 # add latest changes in a git repository 
 git add . && git commit -m "feat: implemented upload functionality in GraphQL API"
```


Next, create a public repository on GitHub using the [Github guide](https://docs.github.com/en/github/getting-started-with-github/create-a-repo) and push the local source code into your new repository.

From your Digitalocean dashboard, navigate to the Apps section and select GitHub as the source to connect your Digitalocean account to GitHub. After the integration, select the newly created repository above from the Repository dropdown. 

In the next configuration page, define the environment variables for the application as defined in your local `.env` file as shown below;



<p id="gdcalert5" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image5.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert6">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image5.png "image_tooltip")


Leaving other settings at their defaults, click the **Next** button to move to the next page where you would give this deployment a unique name, then navigate to the remaining pages to finalize the deployment and build the app.


### Step 6 — Testing The Deployed GraphQL Endpoint

At this point, the application has been fully deployed to DigitalOcean, with a healthy running status similar to the one shown in the image below;



<p id="gdcalert6" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image6.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert7">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image6.png "image_tooltip")


Take note of the endpoint URL of your deployed application placed below the application name. you would use this endpoint to test the upload feature implemented in the deployed GraphQL API with Postman as an API testing tool.

**Note**: _If you do not have the Postman Desktop App installed on your local machine, you can make use of the Postman Web Client within your browser._

From your Postman collection, create a new POST request with a form-data body having the following keys; 

Note: You should replace the USER_ID placeholder with the ID of a user you created using the `CreateUser` mutation. You would also need to change the file field below to a file type before you can select a file on your machine.



*   operations: [{"key":"operations","value":"{\"query\": \"mutation uploadImage($userID: String! $image: Upload!) {\\n uploadProfileImage(input: { userId: $userID file: $image }
*   map: {"file": ["variables.image"], "userId": ["variables.userID"]}
*   userId: &lt;USER_ID>
*   file: &lt;LOCAL_FILE>



<p id="gdcalert7" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/image7.png). Store image on your image server and adjust path/filename/extension if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert8">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/image7.png "image_tooltip")


Hit the Send button to send the POST request, then reload your Digitalocean Space bucket to see your newly uploaded file.


### Conclusion

By reading this article, you have learned about Digitalocean Spaces, and how you can perform file uploads to a created bucket on Digitalocean spaces using the AWS SDK for Golang from a mutation resolver in a GraphQL application.

If you would like to learn more about using Digitalocean Spaces, we encourage you to have a look at its [documentation](https://docs.digitalocean.com/products/spaces/), as it contains an explanation of all aspects of Digitalocean Spaces.

 
