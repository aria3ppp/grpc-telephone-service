package gapi_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"

	"github.com/aria3ppp/grpc-telephone-service/gapi"
	"github.com/aria3ppp/grpc-telephone-service/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func setup(ctx context.Context) (client pb.TelephoneClient, teardown func()) {
	listener := bufconn.Listen(101024 * 1024)

	server := grpc.NewServer()
	pb.RegisterTelephoneServer(server, gapi.NewServer())

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal("setup: failed to serve server:", err)
		}
	}()

	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return listener.Dial()
			},
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("setup: failed to dial server:", err)
	}

	// prepare teardown function
	teardown = func() {
		if err := listener.Close(); err != nil {
			log.Fatal("teardown: failed to close listener:", err)
		}
		server.Stop()
	}

	return pb.NewTelephoneClient(conn), teardown
}

func TestGetContact(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, teardown := setup(ctx)
	t.Cleanup(teardown)

	number := "number"

	// not found
	getResponse, err := client.GetContact(
		ctx,
		&pb.GetContactRequest{Number: number},
	)
	require.Equal(status.Error(codes.NotFound, "contact not found"), err)
	require.Nil(getResponse)

	// add contact
	addStream, err := client.AddContact(ctx)
	require.NoError(err)
	err = addStream.Send(&pb.AddContactRequest{Number: number})
	require.NoError(err)
	addResponse, err := addStream.CloseAndRecv()
	require.NoError(err)
	require.True(
		proto.Equal(&pb.AddContactResponse{ContactsCount: 1}, addResponse),
	)

	// get contact
	getResponse, err = client.GetContact(
		ctx,
		&pb.GetContactRequest{Number: number},
	)
	require.NoError(err)
	require.True(
		proto.Equal(&pb.GetContactResponse{Number: number}, getResponse),
	)
}

func TestListContact(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, teardown := setup(ctx)
	t.Cleanup(teardown)

	// first there is no contact
	listStream, err := client.ListContacts(ctx, &pb.ListContactsRequest{})
	require.NoError(err)

	listResponse, err := listStream.Recv()
	require.Equal(io.EOF, err)
	require.Nil(listResponse)

	// add contacts
	contacts := []*pb.ListContactsResponse{
		{
			Name:     "John",
			Lastname: "Doe",
			Number:   "11111111111",
		},
		{
			Name:     "Regina",
			Lastname: "Phalange",
			Number:   "22222222222",
		},
		{
			Name:     "Donnie",
			Lastname: "Darko",
			Number:   "33333333333",
		},
	}
	addStream, err := client.AddContact(ctx)
	require.NoError(err)
	for _, c := range contacts {
		err = addStream.Send(&pb.AddContactRequest{
			Name:     c.Name,
			Lastname: c.Lastname,
			Number:   c.Number,
		})
		require.NoError(err)
	}
	addResponse, err := addStream.CloseAndRecv()
	require.NoError(err)
	require.True(
		proto.Equal(
			&pb.AddContactResponse{ContactsCount: int32(len(contacts))},
			addResponse,
		),
	)

	// list contacts
	listStream, err = client.ListContacts(ctx, &pb.ListContactsRequest{})
	require.NoError(err)

	for _, c := range contacts {
		listResponse, err = listStream.Recv()
		require.NoError(err)
		require.True(proto.Equal(c, listResponse))
	}
	listResponse, err = listStream.Recv()
	require.Equal(io.EOF, err)
	require.Nil(listResponse)
}

func TestAddContact(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, teardown := setup(ctx)
	t.Cleanup(teardown)

	// add contacts
	contacts := []*pb.ListContactsResponse{
		{
			Name:     "John",
			Lastname: "Doe",
			Number:   "11111111111",
		},
		{
			Name:     "Regina",
			Lastname: "Phalange",
			Number:   "22222222222",
		},
		{
			Name:     "Donnie",
			Lastname: "Darko",
			Number:   "33333333333",
		},
	}
	addStream, err := client.AddContact(ctx)
	require.NoError(err)
	for _, c := range contacts {
		err = addStream.Send(&pb.AddContactRequest{
			Name:     c.Name,
			Lastname: c.Lastname,
			Number:   c.Number,
		})
		require.NoError(err)
	}
	addResponse, err := addStream.CloseAndRecv()
	require.NoError(err)
	require.True(
		proto.Equal(
			&pb.AddContactResponse{ContactsCount: int32(len(contacts))},
			addResponse,
		),
	)

	// make sure contacts are added
	listStream, err := client.ListContacts(ctx, &pb.ListContactsRequest{})
	require.NoError(err)

	for _, c := range contacts {
		listResponse, err := listStream.Recv()
		require.NoError(err)
		require.True(proto.Equal(c, listResponse))
	}
	listResponse, err := listStream.Recv()
	require.Equal(io.EOF, err)
	require.Nil(listResponse)
}

func TestSendMessage(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, teardown := setup(ctx)
	t.Cleanup(teardown)

	stream, err := client.SendMessage(ctx)
	require.NoError(err)

	// say hi
	err = stream.Send(&pb.SendMessageRequest{Msg: "Hi!"})
	require.NoError(err)

	// expect hello
	sendResponse, err := stream.Recv()
	require.NoError(err)
	require.True(
		proto.Equal(&pb.SendMessageResponse{Msg: "Hello!"}, sendResponse),
	)

	for i := 0; i < 10; i++ {
		// say anything else
		err = stream.Send(&pb.SendMessageRequest{Msg: "Anything else!"})
		require.NoError(err)

		// expect sorry
		sendResponse, err := stream.Recv()
		require.NoError(err)
		require.True(
			proto.Equal(
				&pb.SendMessageResponse{Msg: "Sorry, I don't understand :/"},
				sendResponse,
			),
		)
	}

	// close sending end
	err = stream.CloseSend()
	require.NoError(err)

	// expect EOF
	sendResponse, err = stream.Recv()
	require.Equal(io.EOF, err)
	require.Nil(sendResponse)
}
