package remote

// Copyright (C) Philip Schlump 2014.
// MIT Licensed

import (
	"reflect"
	"testing"
)

func Test_parseEndpointUsers(t *testing.T) {
	// func parseEndpointUsers(endpoints []string) (returnEndpoints []string, username, password string, err error) {
	tests := []struct {
		endpointsRaw    []string
		expectEndpoints []string
		expectUsername  string
		expectPassword  string
		expectError     bool
	}{
		{
			endpointsRaw:    []string{"127.0.0.1:2370", "127.0.0.1:2371"},
			expectEndpoints: []string{"127.0.0.1:2370", "127.0.0.1:2371"},
		},
		{
			endpointsRaw:    []string{"etcd://bob:secret@127.0.0.1:2370", "etcd://bob:secret@127.0.0.1:2371"},
			expectEndpoints: []string{"127.0.0.1:2370", "127.0.0.1:2371"},
			expectUsername:  "bob",
			expectPassword:  "secret",
		},
		{
			endpointsRaw:    []string{"127.0.0.1:2370"},
			expectEndpoints: []string{"127.0.0.1:2370"},
		},
		{
			endpointsRaw:    []string{"etcd://bob:secret@127.0.0.1:2370"},
			expectEndpoints: []string{"127.0.0.1:2370"},
			expectUsername:  "bob",
			expectPassword:  "secret",
		},
	}

	for ii, test := range tests {
		gotEndpoints, gotUsername, gotPassword, gotErr := parseEndpointUsers(test.endpointsRaw)
		if test.expectError && gotErr == nil {
			t.Errorf("Error %2d, expectd an error, got none", ii)
		}
		if !test.expectError && gotErr != nil {
			t.Errorf("Error %2d, expectd to have success, got error:%s", ii, gotErr)
		}
		if gotUsername != test.expectUsername {
			t.Errorf("Error %2d, expectd %v - got %v\n", ii, test.expectUsername, gotUsername)
		}
		if gotPassword != test.expectPassword {
			t.Errorf("Error %2d, expectd %v - got %v\n", ii, test.expectPassword, gotPassword)
		}
		if !reflect.DeepEqual(gotEndpoints, test.expectEndpoints) {
			t.Errorf("Error %2d, expectd %v - got %v\n", ii, test.expectEndpoints, gotEndpoints)
		}
	}

}
