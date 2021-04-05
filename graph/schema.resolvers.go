package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/satori/go.uuid"
	"github.com/vickywane/api/graph/generated"
	"github.com/vickywane/api/graph/model"
)

type S3PutObjectAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	user := model.User{
		ID:          fmt.Sprintf("%v", uuid.NewV4()),
		FullName:    input.FullName,
		Password:    input.Password,
		Email:       input.Email,
		ImgURI:      "",
		DateCreated: time.Now().Format("01-02-2006"),
	}

	if err := r.DB.Insert(&user); err != nil {
		return nil, fmt.Errorf("error inserting user: %v", err)
	}

	return &user, nil
}

func PutFile(c context.Context, api S3PutObjectAPI, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return api.PutObject(c, input)
}

func (r *mutationResolver) UploadProfileImage(ctx context.Context, input model.ProfileImage) (bool, error) {
	SpaceName := os.Getenv("DO_SPACE_NAME")
	key := os.Getenv("ACCESS_KEY")
	secret := os.Getenv("ACCESS_SECRET")
	token := os.Getenv("API_TOKEN")

	_, userErr := r.GetUserField("id", *input.UserID); if userErr != nil {
		fmt.Errorf("error getting user: %v", userErr)
	}

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			SigningName: "digitaloceanspaces",
			URL:         fmt.Sprintf("https://.fra1.digitaloceanspaces.com"),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		func(options *config.LoadOptions) error {
			options.Credentials = credentials.NewStaticCredentialsProvider(key, secret, token)
			options.EndpointResolver = customResolver
			options.Region = "fra1"

			return nil
		},
	); if err != nil {
		fmt.Errorf("error getting config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(SpaceName),
	})

	fmt.Printf("Error: %v")
	fmt.Println(output)
	//objectsInput := &s3.ListObjectsV2Input{
	//	Bucket:  aws.String(SpaceName),
	//}

	//objects, err := client.ListObjectsV2(context.TODO(), objectsInput)
	//
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println(objects)

	//fileInput := &s3.PutObjectInput{
	//	Bucket: aws.String(SpaceName),
	//	Key:    aws.String(input.File.Filename),
	//	Body:   input.File.File,
	//	ACL:    "public-read",
	//}
	//
	////_, putErr := client.PutObject(context.TODO(), fileInput); if putErr != nil {
	////	fmt.Printf("error uploading file: %v", err)
	////}

	return true, nil
}

func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	var users []*model.User

	r.DB.Model(&users).Select()

	return users, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
