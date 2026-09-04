package basehdlr

import (
	"context"
	"fmt"

	grpc "bitbucket.org/theruister/goBase/pkg/api/v1/gen/go"
)

func main() {
	fmt.Println("vim-go")
}

func AddUser(ctx context.Context, user *grpc.AddUserRequest) (*grpc.AddUserResponse, error) {
	var ret grpc.AddUserResponse
	ret.UserId = "testId"
	ret.UserName = "testName"

	return &ret, nil
}

func GetUsers(ctx context.Context, user *grpc.AddUserRequest) (*grpc.AddUserResponse, error) {
	var ret grpc.AddUserResponse
	ret.UserId = "testId"
	ret.UserName = "testName"

	return &ret, nil
}
