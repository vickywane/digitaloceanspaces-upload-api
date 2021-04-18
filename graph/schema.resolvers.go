package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"

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
		ImgURI:      "",
		DateCreated: time.Now().Format("01-02-2006"),
	}

	if err := r.DB.Insert(&user); err != nil {
		return nil, fmt.Errorf("error inserting user: %v", err)
	}

	return &user, nil
}

func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	panic("not done")
}

func (r *mutationResolver) UploadProfileImage(ctx context.Context, input model.ProfileImage) (bool, error) {
	SpaceName := os.Getenv("DO_SPACE_NAME")
	SpaceRegion := os.Getenv("DO_SPACE_REGION")
	key := os.Getenv("ACCESS_KEY")
	secret := os.Getenv("ACCESS_SECRET")

	_, userErr := r.GetUserField("id", *input.UserID); if userErr != nil {
		fmt.Errorf("error getting user: %v", userErr)
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
		fmt.Printf("Error opening file: %v", openErr)
	}

	defer file.Close()

	buffer := make([]byte, input.File.Size)

	file.Read(buffer)

	fileBytes := bytes.NewReader(buffer)

	object := s3.PutObjectInput{
		Bucket: aws.String(SpaceName),
		Key:    aws.String(input.File.Filename),
		Body:   fileBytes,
		ACL:    aws.String("public-read"),
	}

	_, uploadErr := s3Client.PutObject(&object)
	if uploadErr != nil {
		fmt.Printf("error uploading %v", uploadErr)

		return false, nil
	}

	os.Remove("image.png")

	return true, nil
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
