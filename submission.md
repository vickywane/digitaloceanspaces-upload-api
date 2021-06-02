

# Digitalocean Golang file uploads API


### Introduction

The ability of a user to upload their personal files when using an application is often considered as a needed feature. However, when using a GraphQL API, this feature could become a challenge to implement, especially with GraphQL’s single source of truth design in your client application.

In this article, you would build a GraphQL API using Golang that has the ability to recieve a media file from a [multipart HTTP request](https://swagger.io/docs/specification/describing-request-body/multipart-requests/) and upload the file to a bucket within [Digitalocean Spaces](https://www.digitalocean.com/products/spaces/).

As you follow through the steps in this article, you would learn about the [Spaces](https://www.digitalocean.com/products/spaces/), and [Managed Databases](https://www.digitalocean.com/products/managed-databases/) products from [Digitalocean](https://www.digitalocean.com) and how you can programmatically upload files to a created bucket from a Golang Application using an S3-compatible [AWS-GO](https://docs.aws.amazon.com/sdk-for-go/api/) SDK.

## Prerequisites

To get the best out of this article, you would need the following;

-  Basic knowledge of [Golang](https://golang.org/). If you are new to Golang, the [How To Write Your First Program In Go](https://www.digitalocean.com/community/tutorials/how-to-write-your-first-program-in-go) article practically explains the Golang Programming Language and the [How To Code in Go](https://www.digitalocean.com/community/tutorial_series/how-to-code-in-go) series contains articles explaining how to configure a [MacOS](https://www.digitalocean.com/community/tutorials/how-to-install-go-and-set-up-a-local-programming-environment-on-macos), [Linux](https://www.digitalocean.com/community/tutorials/how-to-install-go-and-set-up-a-local-programming-environment-on-ubuntu-18-04), and [Windows](https://www.digitalocean.com/community/tutorials/how-to-install-go-and-set-up-a-local-programming-environment-on-windows-10) computer for building Golang applications.

-   An understanding of [GraphQL](https://graphql.org/). Although the GraphQL terminologies used in this article are explained, the [Introduction To GraphQL](https://www.digitalocean.com/community/conceptual_articles/an-introduction-to-graphql) article gives a deeper introduction into what GraphQL APIs are all about.

-   A [Digitalocean account](https://www.digitalocean.com/), as the [Spaces](https://www.digitalocean.com/products/spaces/) and [App Platform](https://www.digitalocean.com/products/app-platform/) products from Digitalocean are used within this article.

-  [Git](https://git-scm.com/) installed and configured on your local machine.

## Terminologies

Below are some frequently terminologies used when working GraphQL. Understanding what these terminologies mean would be quite helpful as they often used when working within this tutorial;

- Resolver: As the name implies, a resolver is a function that resolves or returns a value for a GraphQL field. This value could be an object or a scalar type such as a string, number or even a boolean. In this article, we would use a resolver to mutate data within the GraphQL API.

- Query:  A query is an operation in GraphQL to fetch data, similar to the `GET` HTTP verb in a REST API.

- Mutation: A mutation is an operation used to insert or mutate exisiting data in a GraphQL application, similar to the `POST`, `PATCH`, `PUT` HTTP verbs in a REST API. 

## Step 1 — Bootstrapping a Golang GraphQL API

In this article, you would use the [Gqlgen](https://github.com/99designs/gqlgen) library to boostrap the GraphQL API. Gqlgen is a Go library for building GraphQL APIs. A Schema first approach and Code generation are two important features which Gqlgen provides that would be beneficial while building this API.

Using the Schema First Approach feature, you get to define the data model for the API using the GraphQL [Schema Definition Language](http://graphql.org/learn/schema/) (SDL), then you generate the boilerplate code for the API from the defined schema usig the code generation feature. 

Execute the command below from your terminal in your project directory to create a `go.mod` file that manages the modules within the API project;

```command 
 go mod init
```

Next, install the Gqlgen library into your project;

```command
 go get github.com/99designs/gqlgen
```

Then using the installed Gqlgen library, generate the boilerplate files needed for a GraphQL API;

```command
 gqlgen init 
```

Running the `gqlgen` command above would generate a `server.go` file for running the GraphQL server and a `graph` directory containing a `schema.graphqls` file that would contain the Schema Definitions for the GraphQL API.


### Defining Application GraphQL Schema

By default, the `gqlgen init` command previously executed would generate the schema for a TODO application within the `schema.graphqls` file. While this is a valid schema, the application intentded for this tutorial is not a TODO application. Hence, you would need to change the boilerplate schema.

To create a suitable schema for the API we are building, open the `schema.graphqls` file in your prefferred code editor and replace the boilerplate schema with the schema in the code snippet below;


```graphql
[label schema.graphls]

scalar Upload

type User {
  id: ID!
  fullName: String!
  email: String!
  img_uri: String!
}

type Query {
  user: User!
}

input NewUser {
  fullName: String!
  email: String!
  img_uri: String
  password : String!
}

input ProfileImage {
  userId: String
  file: Upload
}

type Mutation {
  createUser(input: NewUser!): User!
  uploadProfileImage(input: ProfileImage!): Boolean!
}
```


The code block above contains a schema with three types; the Upload and User types which are known as Object types in the GraphQL [Schema Definition Language](http://graphql.org/learn/schema/).

<$>[note]
**Note:** The Upload scalar type is automatically defined by Gqlgen and it contains the properties of a file. To use it, you only need to declare it at the top of the schema file, as it was done in the code block above.
<$>

The schema in the code block above also contains a Mutation type containing the `CreateUser` and `uploadProfileImage` fields and the `user` field returning a single user type as the `Query`. 

At this point, you have defined the structure of the data model for the application through the `schema.graphq1s` file, the next step is to generate the query and mutation resolvers functions for the schema above using Gqlgen's code generation feature.


## Step 2 — Generating Application Resolvers

The Gqlgen package being used is based on a schema first approach. A time-saving feature of Gqlgen is its ability to generate your application’s resolvers based on your defined schema in the `schema.graphqls` file. With this feature, you do not need to manually write the resolver boilerplate code, all you need to do is to focus on the actual implementation of the defined resolvers.

To utilize the code generation feature, execute the command below from a terminal in the project directory to generate the GraphQL API model files and resolvers;

```command
 gqlgen generate 
```

After executing the gqlgen command above, you would observe that some new files have been generated and your project now has the folder structure shown below;

![Generated Folder Structure](https://i.imgur.com/APlA6d7.png)

Among the files shown in the image above, of interest is the `schema.resolvers.go` file. As shown in the code block below, it contains an implementation of the Mutation and Query types previously defined in the `schema.graphqls` file. 

```go

package graph

import (
	"context"
	"fmt"

	"github.com/vickywane/graphql-api/graph/generated"
	"github.com/vickywane/graphql-api/graph/model"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UploadProfileImage(ctx context.Context, input model.ProfileImage) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

```

As defined in the `schema.graphqls` file, two mutations, and one query resolver functions where generated by Gqlgen's code generator. These resolvers would serve the following purposes;

* CreateUser - This would mutation resolver would be used to insert a new user record into the connected Postgres database.

* UploadProfileImage - This mutation resolver would be used to upload a media file recieved from a [multipart HTTP request](https://swagger.io/docs/specification/describing-request-body/multipart-requests/) and upload the file to a bucket within [Digitalocean Spaces](https://www.digitalocean.com/products/spaces/). After the file upload, an update would be made to insert the url of the uploaded file into the `img_uri` field of the previously created user.

*  Users - This query resolver would query the database for all exisiting users and return them as the query result.

Going through the functions generated from the Mutation and Query types, you would observe that they cause a [panic](https://golang.org/src/runtime/panic.go) with a "**not implemented**" error when executed. This indicates that they are still auto-generated boilerplate code. Later in this tutorial, we would come back to the `schema.resolver.go` file to implement these generated functions.

## Step 3 — Provisioning and Using a Managed Database Instance on DigitalOcean

Although the application would not store images directly in a database, it still needs a database to insert each user‘s record. The stored record would then contain links to the uploaded files.

A user’s record would consist of a **Fullname**, **email**, **dateCreated,** and an **img_uri** field of String data type. The **img_uri** field would contain the URL pointing to an image file uploaded by a user through this GraphQL API and stored within a bucket on Digitalocean spaces.

Using your Digitalocean dashboard, navigate to the Databases section of the console to create a new database cluster. By default, PostgreSQL would be the selected database to run within this cluster. Leave all other settings at their default values and proceed to create this cluster using the button at the bottom.

After the cluster has been created, the connection details of the cluster would be displayed. Create a `.env` file within the GraphQL-API project directory to securely store the cluster credentials as environment variables in the following format;

```bash
[label .env]

 DB_PASSWORD=<PASSWORD>
 DB_PORT=<PORT>
 DB_NAME=<DATABASE>
 DB_ADDR=<HOST>
 DB_USER=<USERNAME>
```


With the connection details securely stored in the .env file above, the next step would be to connect to the database cluster through our backend application.

Create a `db.go` file within the `graph` package directory. We would gradually put together the code within the file to establish a connection with the Postgres database created in the [Managed Databases](https://www.digitalocean.com/products/managed-databases/) cluster.

First, add the content of the code block below into the `db.go` file to create a user table in the Postgres database immediately a connection to the database has been established.

```go
[label server.go]
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
```

Using the `IfNotExists` option passed to `CreateTable` method from [go-pg](https://github.com/go-pg/pg), the `createSchema` function in the code block above only inserts a new table into the database if the table does not exist. You can understand this process as a simplified form of seeding a newly created database, rather than creating the Tables manually through a command line client or GUI, the `createSchema` function takes care of the table creation.

Next, add the content of the code block below into the `db.go` file to establish a connection to the Postgres database and execute the `createSchema` function above when a connection has been established successfully.

```go
[label server.go]

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

The exported `Connect` function in the code block above when executed establishes a connection to a Postgres database using [go-pg](https://github.com/go-pg/pg) and returns the connection instance. This done through the following operations explained below;

* First, the database credentials you stored in the root `.env` file are retrieved, then, a variable is created to store a string formatted with the retrieved credentials. This variable would be used as a connection URI when connecting with database.

* Next, the created connection string is parsed to know if the formatted credentials are valid. If valid, the connection string is passed into the `connect` method as an argument to establish a connection. 

To use the exported `Connect` function, you need to add it to the `server.go` file to execute the `Connect` function when the application is started and the instance would also be available in the `Resolver` struct.

Open the `server.go` file in your preferred code editor and add the lines highlighted below into the `server.go` file to utilize the previously created `db` package immediately after the application is started.


```go
[label db/db.go]
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


``` go
[label resolver.go]
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

#### Mutation Resolvers

Going through the `schema.graphqls` file, there are only two mutation resolvers generated. One with the purpose of handling the user's creation, while the other to handle the profile image uploads.

Modify the `CreateUser` mutation with the code snippet below to insert a new row containing the user details input into the database


```go
[label schema.resolver.go]
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


Going through the `CreateUser` mutation in the code snippet above, you would observe two things about the user rows inserted;

*   Each row inserted is given a unique UUID formatted as a string.
*   The `ImgURI` field in each row has a placeholder image URL as the default value. This would be updated when a user uploads a new image. 

To test the resolver above from your browser, navigate to `http://localhost:8080` to access the GraphQL playground built-in to your GraphQL API. Paste the GraphQL Mutation in the code block below into the playground editor to insert a new user record.

```graphql
[label graphql]

mutation createUser {
  createUser(
    input: {
      email: "johndoe@gmail.com"
      fullName: "John Doe"
      password: "password"
    }
  ) {
    id
  }
}
```

![A create user muation on the GraphQL Playround](https://i.imgur.com/57Q16Ir.png)

Going through the image above, you executed the `CreateUser` mutation to create a test user with the name of **John Doe**, and you returned the `id` of the newly inserted user record as the result of the mutation. 

At this point, you have the second `UploadProfileImage` mutation resolver function left to implement, but before you implement this function, you need to implement the query resolver first. This is because each upload is linked to a specific user, hence the need to retrieve the ID of a specific user before uploading an image.

#### Query Resolver

As defined in the `schema.graphqls` file, one query resolver was generated for the purpose of retrieving all created users.

Modify the generated `Users` query resolver with the code block below to query the postgres database for all user rows. 


```go
[label schema.graphqls]
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

Within the `Users` reswolver function above, fetching all users record is made possible by using go-pg’s `select` method on the `User` model without passing `WHERE` or `LIMIT` clause into the query.

To test this query resolver from your browser, navigate to `http://localhost:8080` to access the GraphQL playground. Paste the GraphQL Query below into the playground editor to fetch all created user record.

```graphql
[label graphql]

query getUsers {
  users {
      email
      fulName
      dateCreated
      id
      img_uri
  }
}
```

![A query muation to retrieve a user using the GraphQL Playround](https://i.imgur.com/WPLxxm7.png)

Going through the right side of the image above, you would notice that a `users` object having an array value was returned. For now only the previously created user was returned in the `users` array because that it is the only record in the table. More users would be returned in the `users` array if you execute the previous Mutation with new user values.

At this point, you have now implemented both the `CreateUser` mutation and the `User` query. Everything is in place for you to implement recieving images from the second `UploadProfileImage` resolver and uploading the recieved image to a bucket with Digitalocean Spaces through the use of an S3 compatible [AWS-GO](https://docs.aws.amazon.com/sdk-for-go/api/) SDK.

## Step 5 — Uploading Images To Digitalocean Spaces

[Spaces](https://www.digitalocean.com/products/spaces/) is a simple, and scalable cloud based object storage service from Digitalocean. You would use the powerful API within the second `UploadProfileImage` mutation to upload images.

To begin, navigate to the Spaces section of your DigitalOcean console where you would create a new bucket for storing the uploaded files from your backend application.

Click the **Create New Space** button, leaving other settings at their default values and specify a unique name for the new space as shown below before creating your new space;

![Digitalocean spaces](https://i.imgur.com/Aifnmzf.png)

After a new Space has been created, navigate to the settings tab and copy the space’s endpoint into the GraphQL project environment variables.


```bash
[label .env]
SPACE_ENDPOINT=<BUCKET_ENDPOINT>
```


Next, using [this guide](https://docs.digitalocean.com/products/spaces/how-to/manage-access/#acwithin the Digitalocean Spaces documentation that explains the process of creating secret and access keys, create a secret and access key for the backend applicationcess-keys) within the Digitalocean Spaces documentation that explains the process of creating secret and access keys, create a secret and access key for the backend application. After creating them, copy the values and store them in your backend application’s `.env` file in the format below;


```bash
[label .env]
ACCESS_KEY=<SPACE_ACCESS_KEY>
SECRET_KEY=<SPACE_SECRET_KEY>
```


One way to perform operations on your Digitalocean Space is through the use of supported AWS SDKs as they are quite compatible. The Digitalocean Spaces [documentation](https://docs.digitalocean.com/products/spaces/) provides a list of operations you can perform on the Spaces API using AWS SDK.

To implement the file uploads, modify the `UploadProfileImage` mutation function within the `Schema.resolvers.go` file with the code below which uploads an image from the resolver function into our Digitalocean Spaces bucket.


```go
[label schema.resolvers.go]
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


Going through the entire function above, you would observe the following operations being performed in the following order;


*   First, using the helper function in the `resolver.go` file, the user row having the UserID argument as an ID is queried from the database to confirm that you are trying to upload a file for an actual user.
*   Next, you configured the SDK client for Digitalocean spaces using an access key and secret key credentials obtained from the Digitalocean console.
*   Next, using the `ReadAll` method from the `io` package, you read the entire content of the file added to the HTTP request sent to the GraphQL API, then a temporary file was created to dump this content into.
*   Next, you added the `PutObjectInput` struct fields having the `Bucket` field as the name of the Space on DigitalOcean, the `Key` field as the name of the file being uploaded, the `Body` field as the temporary file you created, and lastly the ACL ( Access Control Lock) field to set the permission type on the file.

<$>[note]
**Note:** The `public-read` ACL type is used because we want all uploaded files to be public and open to anyone with the link for viewing.
<$>

*   Lastly, after the file is uploaded, the temporary file is deleted, and a link to the uploaded file is formatted together using the Spaces endpoint and filename, then the user’s ImgUri field in the row is updated to contain the formatted link.

<$>[warning]
The reason why a temporary file is used to store the file uploaded file is that when using the AWS v1 SDK for Golang, the `body` property within the `PutObjectInput` struct has a `Reader.seek` type. You should put this into consideration if you expect your users to upload files with huge sizes.
<$>

To test the new mutation resolver, execute the command below to make an HTTP request to the GraphQL API using cURL, adding an image into the request form body.


``` command
curl localhost:8080/query  -F operations='{ "query": "mutation uploadProfileImage($image: Upload! $userId : String!) { uploadProfileImage(input: { file: $image  userId : $userId}) }", "variables": { "image": null, "userId" : "121212" } }' -F map='{ "0": ["variables.image"] }'  -F 0=@sample.jpeg
```

After the file has been uploaded, the following boolean status would be printed out in the terminal as the request-response, indicating that the file upload was successful.


```
[secondary_label Output]
{"data": { "uploadProfileImage": true }}
```

Going through your created bucket within the Spaces section of the Digitalocean console, you would find the image recently uploaded from your terminal.

![A bucket within Digitalocean showing a list of uploaded files](https://i.imgur.com/o4f5P7N.png)


Also, if you query the user’s data after a successful file upload, you would observe that the img_uri field returned in the user’s data points to the file recently uploaded to your bucket.


At this point, you have a functional backend application exposing a GraphQL with mutation resolvers that you can use to insert a new user record into a connected Postgresql database and also upload an image for the new user.

you can move a step further to deploy this application to the Digitalocean App platform. Using the App Platform’s support for Golang, a deployment only requires a minimal configuration.


### Deploying GraphQL API to App Platform
<$>[info]
**Info:**  [App platform](https://www.digitalocean.com/products/app-platform/) is a Digitalocean service product that makes it much easier to build, deploy, and even scale your applications. App platform supports a variety of languages and within this article, you would utilize the support for applications written in Go and stored within GitHub.

To begin depoying your backend application, create a local git repository by executing the command below from a terminal;
<$>

```command
 git init
```

Add your latest changes from all files into the local repository by running the command below from your terminal

```command 
 git add .
```

Then using the command below, commit the recent file changes made within the repository.

```command
git commit -m "feat: implemented upload functionality"
```


Next, create a public repository on GitHub using the [Github guide](https://docs.github.com/en/github/getting-started-with-github/create-a-repo) and push the local source code into your new repository.

From your Digitalocean dashboard, navigate to the Apps section and select GitHub as the source to connect your Digitalocean account to GitHub. After the integration, select the newly created repository above from the Repository dropdown.

In the next configuration page, define the environment variables for the application as defined in your local `.env` file as shown below;


![Environment variables for deploying a Golang Application to Digitalocean App Platform](https://i.imgur.com/tiEl0wx.png)


Leaving other settings at their defaults, click the **Next** button to move to the next page where you would give this deployment a unique name, then navigate to the remaining pages to finalize the deployment and build the app.


### Step 6 — Testing The Deployed GraphQL Endpoint

At this point, the application has been fully deployed to DigitalOcean, with a healthy running status similar to the one shown in the image below;

![Health status of a Golang application deployed to Digitalocean App Platform](https://i.imgur.com/Nbjeph7.png)


Take note of the endpoint URL of your deployed application placed below the application name. you would use this endpoint to test the upload feature implemented in the deployed GraphQL API with Postman as an API testing tool.

<$>[info]
**Note:** If you do not have the Postman Desktop App installed on your local machine, you can make use of the Postman Web Client within your browser._
<$>

From your Postman collection, create a new POST request with a form-data body having the following keys;

<$>[info]
**Note**: You should replace the USER_ID placeholder with the ID of a user you created using the `CreateUser` mutation. You would also need to change the file field below to a file type before you can select a file on your machine.
<$>


*   operations: [{"key":"operations","value":"{\"query\": \"mutation uploadImage($userID: String! $image: Upload!) {\\n uploadProfileImage(input: { userId: $userID file: $image }
*   map: {"file": ["variables.image"], "userId": ["variables.userID"]}
*   userId: &lt;USER_ID>
*   file: &lt;LOCAL_FILE>

![Using Postman Form-Body to add images to a POST request using a GraphQL API](https://i.imgur.com/fOj229w.png)


Hit the Send button to send the POST request, then reload your Digitalocean Space bucket to see your newly uploaded file.


## Conclusion

By reading this article, you have learned about Digitalocean Spaces, and how you can perform file uploads to a created bucket on Digitalocean spaces using the AWS SDK for Golang from a mutation resolver in a GraphQL application.

If you would like to learn more about using Digitalocean Spaces, we encourage you to have a look at its [documentation](https://docs.digitalocean.com/products/spaces/), as it contains an explanation of all aspects of Digitalocean Spaces.