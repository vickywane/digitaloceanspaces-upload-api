# How To Build a GraphQL API With Golang to Upload Files to DigitalOcean Spaces 🚀
This repository contains the completed GraphQL application built in my technical article published on Digitalocean's Developer blog.
Built with ❤️ by [Victory Nwani](https://www.linkedin.com/in/victory-nwani-b820b2157/) ( Send me a message if you're hiring 😃 )

## Needed Stuff

To run this project locally, you will need a few things: 
- The Golang compiler installed on your computer. ( Not strange, this is a Go application. What do you expect?  😐 )
- The Golang dependencies installed. Run `go mod tidy` and you'll be fine! 😃
- Connection details for a PostgreSQL database. This application stores user data in a PostgreSQL DB.

### Database Connection Details:
You'll get a connection error if you skip this part.
The tutorial on Digitalocean explains how to get the details used below. Go read it 😛!

```bash
DB_USER=""
DB_PASSWORD=""
DB_ADDR=""
DB_PORT=25060
DB_NAME=""

SPACE_ENDPOINT=""
DO_SPACE_REGION=""
DO_SPACE_NAME=""
ACCESS_KEY=""
SECRET_KEY=""
```

## Usage(s) 🛠

You can execute the following GraphQL operations below through the GraphiQL playground at http://localhost:8080

### Testing the `createUser` Mutation Resolver 

This should come first while testing, as it puts some data into the PostgreSQL database
Copy and paste the mutation below to insert John's data 👨. Feel free to change the data to your taste 😉

```gql
    mutation createUser {
      createUser(
        input: {
          email: "johndoe@gmail.com"
          fullName: "John Doe"
        }
      ) {
        id
        fullname
        email
      }
    }
```

### Testing the `fetchUsers` Query Resolver 

This should come next, to enable you view the data inserted through the mutation above. 
Copy and paste the query below into the playground to retrieve all data:

```gql
    query fetchUsers {
      users {
          fullName
          id
          img_uri
      }
    }
```


## Testing the file upload Mutation Resolver
This one is gonna be a tad different as the GraphiQL playground has no support for file uploads.
You can however try using Postman or Insomia's support for file GraphQL. Rumour has it that they support file uploads laugh 😂. 

Execute the command below from a terminal with `curl` installed to make a http request containing an image in the form body.

**Note:** Execute the command below from this project's root directory to use the `sample.jpeg` image as a test file.

```
bash 
curl localhost:8080/query  -F operations='{ "query": "mutation uploadProfileImage($image: Upload! $userId : String!) { uploadProfileImage(input: { file: $image  userId : $userId}) }", "variables": { "image": null, "userId" : "121212" } }' -F map='{ "0": ["variables.image"] }'  -F 0=@sample.jpeg
```

## One last thing 🤫 

Please star ( ⭐️ ) this repository if you find this application useful. The stars ( 🌟 ) provide encouragement. 

Happy Hacking ✌️