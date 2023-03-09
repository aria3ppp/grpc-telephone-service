package gapi

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aria3ppp/grpc-telephone-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TelephoneServer struct {
	pb.UnsafeTelephoneServer
	contacts []*pb.ListContactsResponse
}

func NewServer() *TelephoneServer {
	s := &TelephoneServer{}
	return s
}

func (s *TelephoneServer) GetContact(
	ctx context.Context,
	req *pb.GetContactRequest,
) (*pb.GetContactResponse, error) {
	for _, c := range s.contacts {
		if req.Number == c.Number {
			return &pb.GetContactResponse{
				Name:     c.Name,
				Lastname: c.Lastname,
				Number:   c.Number,
			}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "contact not found")
}

func (s *TelephoneServer) ListContacts(
	_ *pb.ListContactsRequest,
	stream pb.Telephone_ListContactsServer,
) error {
	for _, c := range s.contacts {
		err := stream.Send(&pb.ListContactsResponse{
			Name:     c.Name,
			Lastname: c.Lastname,
			Number:   c.Number,
		})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func (s *TelephoneServer) AddContact(
	stream pb.Telephone_AddContactServer,
) error {
	contactCount := 0

	for {
		contact, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return stream.SendAndClose(&pb.AddContactResponse{
					ContactsCount: int32(contactCount),
				})
			}
			return status.Error(codes.Internal, err.Error())
		}

		s.contacts = append(s.contacts, &pb.ListContactsResponse{
			Name:     contact.Name,
			Lastname: contact.Lastname,
			Number:   contact.Number,
		})

		contactCount++
	}
}

func (s *TelephoneServer) SendMessage(
	stream pb.Telephone_SendMessageServer,
) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Error(codes.Internal, err.Error())
		}

		message := strings.ToLower(req.Msg)
		var response string

		switch message {
		case "hi!":
			response = "Hello!"
		case "how are you?":
			response = "I'm fine!"
		case "see you later":
			response = "See you!"
		default:
			response = "Sorry, I don't understand :/"
		}

		err = stream.Send(&pb.SendMessageResponse{
			Msg: response,
		})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}
