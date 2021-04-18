# GraphQL Upload API


## Testing the file upload mutation resolver
Execute the command below from a terminal with `curl` installed to make a http request containing an image in the form body.

**Note:** Execute the command below from this project's root directory to use the `sample.jpeg` image as a test file.


```
bash 

curl localhost:8080/query  -F operations='{ "query": "mutation uploadProfileImage($image: Upload! $userId : String!) { uploadProfileImage(input: { file: $image  userId : $userId}) }", "variables": { "image": null, "userId" : "121212" } }' -F map='{ "0": ["variables.image"] }'  -F 0=@sample.jpeg

```