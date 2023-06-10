package repository

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "headscale-panel/gen/headscale/v1"
	"testing"
	"time"
)

func TestHeascaleCheckSync(t *testing.T) {
	testOldData := make([]*pb.User, 0)
	//testOldData := []*pb.User{
	//	&pb.User{
	//		Id:        "1",
	//		Name:      "1",
	//		CreatedAt: timestamppb.New(time.Date(2023, 1, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "2",
	//		Name:      "2",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "3",
	//		Name:      "3",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 15, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "4",
	//		Name:      "4",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "6",
	//		Name:      "6",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "7",
	//		Name:      "7",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "8",
	//		Name:      "8",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "9",
	//		Name:      "9",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "10",
	//		Name:      "10",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "11",
	//		Name:      "11",
	//		CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//	&pb.User{
	//		Id:        "12",
	//		Name:      "12",
	//		CreatedAt: timestamppb.New(time.Date(2023, 3, 10, 13, 29, 45, 0, time.UTC)),
	//	},
	//}

	testNewData := []*pb.User{
		&pb.User{
			Id:        "1",
			Name:      "1",
			CreatedAt: timestamppb.New(time.Date(2023, 1, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "2",
			Name:      "2",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "3",
			Name:      "3",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 15, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "7",
			Name:      "7",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "8",
			Name:      "8",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "9",
			Name:      "9",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "10",
			Name:      "10",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "11",
			Name:      "11",
			CreatedAt: timestamppb.New(time.Now()),
		},
		//&pb.User{
		//	Id:        "12",
		//	Name:      "12",
		//	CreatedAt: timestamppb.New(time.Now()),
		//},
		//&pb.User{
		//	Id:        "13",
		//	Name:      "13",
		//	CreatedAt: timestamppb.New(time.Now()),
		//},
		//&pb.User{
		//	Id:        "14",
		//	Name:      "14",
		//	CreatedAt: timestamppb.New(time.Now()),
		//},
		//&pb.User{
		//	Id:        "15",
		//	Name:      "15",
		//	CreatedAt: timestamppb.New(time.Now()),
		//},
	}

	create, delete, history := checkSyncUser(testNewData, testOldData)
	fmt.Println("--------------------------- create users ---------------------------")
	fmt.Println(create)
	fmt.Println("--------------------------- delete users ---------------------------")
	fmt.Println(delete)
	fmt.Println("--------------------------- history users ---------------------------")
	fmt.Println(history)
}

func TestSearchUser(t *testing.T) {
	testNewData := []*pb.User{
		&pb.User{
			Id:        "1",
			Name:      "1",
			CreatedAt: timestamppb.New(time.Date(2023, 1, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "2",
			Name:      "2",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "3",
			Name:      "3",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 15, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "7",
			Name:      "7",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "8",
			Name:      "8",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "9",
			Name:      "9",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "10",
			Name:      "10",
			CreatedAt: timestamppb.New(time.Date(2023, 2, 10, 13, 29, 45, 0, time.UTC)),
		},
		&pb.User{
			Id:        "11",
			Name:      "11",
			CreatedAt: timestamppb.New(time.Now()),
		},
		&pb.User{
			Id:        "12",
			Name:      "12",
			CreatedAt: timestamppb.New(time.Now()),
		},
		&pb.User{
			Id:        "13",
			Name:      "13",
			CreatedAt: timestamppb.New(time.Now()),
		},
		&pb.User{
			Id:        "14",
			Name:      "14",
			CreatedAt: timestamppb.New(time.Now()),
		},
		&pb.User{
			Id:        "15",
			Name:      "15",
			CreatedAt: timestamppb.New(time.Now()),
		},
	}

	data := searchUser(testNewData, "7")
	if data == nil {
		panic(errors.New("not find"))
	}
	fmt.Println(data)

	data = searchUser(testNewData, "5")
	if data != nil {
		fmt.Println(data)
		panic(errors.New("found data"))
	}
}
